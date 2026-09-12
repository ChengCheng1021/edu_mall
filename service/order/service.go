package order

import (
	"context"
	"encoding/json"
	"mall/adaptor"
	"mall/adaptor/redis"
	"mall/adaptor/repo/admin"
	"mall/adaptor/repo/goods"
	"mall/adaptor/repo/model"
	"mall/adaptor/repo/order"
	userRepo "mall/adaptor/repo/user"
	"mall/adaptor/rpc"
	"mall/common"
	"mall/config"
	"mall/consts"
	"mall/service/do"
	"mall/service/dto"
	"mall/utils/logger"
	"mall/utils/pool"
	"mall/utils/tools"

	"github.com/jinzhu/copier"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type Service struct {
	conf       *config.Config
	course     goods.ICourse
	rdsOrder   redis.IOrder
	order      order.IOrder
	aliPay     rpc.IPay
	idNode     redis.IGenID
	adminUser  admin.IAdminUser
	user       userRepo.IUser
	storage    rpc.IStorage
	userCourse userRepo.IUserCourse
}

func NewService(adaptor adaptor.IAdaptor) *Service {
	return &Service{
		conf:       adaptor.GetConfig(),
		course:     goods.NewCourse(adaptor),
		rdsOrder:   redis.NewOrder(adaptor),
		order:      order.NewOrder(adaptor),
		aliPay:     rpc.NewAliPay(adaptor),
		idNode:     redis.NewGenIdNode(adaptor),
		adminUser:  admin.NewAdminUser(adaptor),
		user:       userRepo.NewUser(adaptor),
		storage:    rpc.NewStorage(adaptor),
		userCourse: userRepo.NewUserCourse(adaptor),
	}
}
func (s *Service) GetOrderInfo(ctx context.Context, user *common.UserInfo, req *dto.GetOrderInfoReq) (*dto.OrderInfoResp, common.Errno) {
	tempOrder, err := s.order.GetOrderByID(ctx, req.OrderID)
	if err != nil {
		logger.Error("GetOrderInfo GetOrderByID error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	if user != nil && tempOrder.UserID != user.User.ID {
		return nil, common.OK
	}
	var (
		items   []*model.OrderItem
		refunds []*model.OrderRefund
	)
	tempPool := pool.NewPoolWithSize(2)
	defer tempPool.Release()
	tempPool.RunGo(func() {
		temp, err := s.order.GetOrderItems(ctx, req.OrderID)
		if err != nil {
			logger.Error("GetOrderInfo GetOrderItems error", zap.Error(err), zap.Any("req", req))
			return
		}
		items = temp
	})
	tempPool.RunGo(func() {
		temp, err := s.order.GetOrderRefunds(ctx, req.OrderID)
		if err != nil {
			logger.Error("GetOrderInfo GetOrderRefunds error", zap.Error(err), zap.Any("req", req))
			return
		}
		refunds = temp
	})
	tempPool.Wait()

	tempList := s.convertModelOrderToOrderDto(ctx, []*model.Order{tempOrder})
	if len(tempList) == 0 {
		logger.Error("GetOrderInfo convertModelOrderToOrderDto error", zap.Any("req", req))
		return nil, common.ServerErr.WithMsg("convertModelOrderToOrderDto error")
	}

	orderDto := tempList[0]
	orderItems := make([]*dto.OrderItemDto, 0)
	orderRefunds := make([]*dto.RefundDto, 0)
	itemMap := lo.SliceToMap(items, func(item *model.OrderItem) (int64, *model.OrderItem) {
		return item.ID, item
	})
	itemRefundMap := make(map[int64]*model.OrderRefund)
	refundItemsMap := make(map[int64][]int64)

	copier.Copy(&orderItems, items)
	copier.Copy(&orderRefunds, refunds)
	for index, refund := range refunds {
		itemIds := make([]int64, 0)
		if err := json.Unmarshal([]byte(refund.ItemIds), &itemIds); err != nil {
			logger.Error("GetOrderInfo refund item ids unmarshal error", zap.Error(err), zap.Int64("refund_id", refund.ID))
			continue
		}
		refundItemsMap[refund.ID] = itemIds
		orderRefunds[index].ItemIds = itemIds
		orderRefunds[index].Reason = refund.Reason
		orderRefunds[index].RefundID = refund.RefundID
		orderRefunds[index].ApplyUserID = refund.ApplyUserID
		orderRefunds[index].Status = refund.Status
		//if refund.ApplyUserID > 0 {
		//	orderRefunds[index].RefundName = s.getAdminUserName(ctx, refund.ApplyUserID)
		//}
		for _, itemID := range itemIds {
			itemRefundMap[itemID] = refund
		}
	}

	fileKeys := make([]string, 0)
	goodsSnapMap := make(map[int64]dto.OrderGoodsSnap)
	for _, item := range orderItems {
		orderItem, ok := itemMap[item.ID]
		if !ok {
			logger.Error("GetOrderInfo GetOrderItems error", zap.Any("req", req))
			return nil, common.ServerErr.WithMsg("GetOrderItems error")
		}
		goodsSnap := &dto.OrderGoodsSnap{}
		err = json.Unmarshal([]byte(orderItem.GoodsSnap), goodsSnap)
		if err != nil {
			logger.Error("GetOrderInfo json.Unmarshal error", zap.Error(err), zap.Any("orderItem", orderItem))
			return nil, common.ServerErr.WithMsg("json.Unmarshal error")
		}
		if refund, ok := itemRefundMap[item.ID]; ok {
			item.RefundStatus = refund.Status
		}
		fileKeys = append(fileKeys, goodsSnap.CoverKey, goodsSnap.DetailCoverKey)
		goodsSnapMap[item.ID] = *goodsSnap
	}

	fileUrlMap, err := s.storage.GetPreviewUrl(ctx, &do.GetPreviewUrl{
		Keys:        fileKeys,
		ExpireHours: 6,
	})
	if err != nil {
		logger.Error("GetOrderInfo GetPreviewUrl error", zap.Error(err), zap.Any("req", req))
		return nil, common.ServerErr.WithMsg("GetPreviewUrl error")
	}
	orderItemDtoMap := make(map[int64]*dto.OrderItemDto)
	for _, item := range orderItems {
		goodsSnap := goodsSnapMap[item.ID]
		goodsSnap.CoverUrl = fileUrlMap[goodsSnap.CoverKey]
		goodsSnap.DetailUrl = fileUrlMap[goodsSnap.DetailCoverKey]
		item.GoodsSnap = goodsSnap
		orderItemDtoMap[item.ID] = item
	}
	for index := range orderRefunds {
		itemDtos := make([]*dto.OrderItemDto, 0)
		for _, itemID := range orderRefunds[index].ItemIds {
			if itemDto, ok := orderItemDtoMap[itemID]; ok {
				itemDtos = append(itemDtos, itemDto)
			}
		}
		orderRefunds[index].Items = itemDtos
	}
	resp := &dto.OrderInfoResp{
		OrderDto: orderDto,
		Items:    orderItems,
		Refunds:  orderRefunds,
	}
	return resp, common.OK
}
func (s *Service) convertModelOrderToOrderDto(ctx context.Context, list []*model.Order) []*dto.OrderDto {
	var (
		adminUserIDs []int64
		userIDs      []int64
	)
	lo.ForEach(list, func(item *model.Order, index int) {
		userIDs = append(userIDs, item.UserID, item.CreateBy)
		if item.CancelBy != nil && item.CancelType != nil {
			switch *item.CancelType {
			case consts.AdminUser:
				adminUserIDs = append(adminUserIDs, *item.CancelBy)
			case consts.CustomerUser:
				userIDs = append(userIDs, *item.CancelBy)
			}
		}
	})
	userIDs = lo.Uniq(userIDs)
	adminUserIDs = lo.Uniq(adminUserIDs)

	var (
		adminUserMap    = make(map[int64]string)
		customerNameMap = make(map[int64]string)
		customerMobile  = make(map[int64]string)
	)
	tempPool := pool.NewPoolWithSize(2)
	defer tempPool.Release()
	tempPool.RunGo(func() {
		tempMap, err := s.adminUser.GetUserNameMap(ctx, adminUserIDs)
		if err != nil {
			logger.Error("convertModelOrderToOrderDto GetUserNameMap admin error", zap.Error(err), zap.Any("adminUserIDs", adminUserIDs))
			return
		}
		adminUserMap = tempMap
	})
	tempPool.RunGo(func() {
		if len(userIDs) == 0 {
			return
		}
		nameMap, err := s.user.GetUserNameMap(ctx, userIDs)
		if err != nil {
			logger.Error("convertModelOrderToOrderDto GetUserNameMap customer error", zap.Error(err), zap.Any("userIDs", userIDs))
			return
		}
		customerNameMap = nameMap
		mobileUsers, err := s.user.GetMobileUsersByUserIds(ctx, userIDs)
		if err != nil {
			logger.Error("convertModelOrderToOrderDto GetMobileUsersByUserIds error", zap.Error(err), zap.Any("userIDs", userIDs))
			return
		}
		for _, mobileUser := range mobileUsers {
			mobile, err := tools.AESDecrypt(mobileUser.MobileAes, []byte(s.conf.BizConf.MobileSecret))
			if err != nil {
				logger.Error("convertModelOrderToOrderDto AESDecrypt mobile error", zap.Error(err), zap.Int64("user_id", mobileUser.UserID))
				continue
			}
			customerMobile[mobileUser.UserID] = string(mobile)
		}
	})
	tempPool.Wait()

	retList := make([]*dto.OrderDto, 0)
	copier.Copy(&retList, list)
	lo.ForEach(retList, func(item *dto.OrderDto, index int) {
		item.CreateName = customerNameMap[item.CreateBy]
		if item.CancelType == consts.AdminUser {
			item.CancelName = adminUserMap[item.CancelBy]
		} else {
			item.CancelName = customerNameMap[item.CancelBy]
		}
		item.UserName = customerNameMap[item.UserID]
		item.UserMobile = customerMobile[item.UserID]
	})
	return retList
}
