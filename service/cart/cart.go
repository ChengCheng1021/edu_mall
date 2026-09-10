package cart

import (
	"context"
	"encoding/json"
	"mall/adaptor/repo/model"
	"mall/common"
	"mall/service/do"
	"mall/service/dto"
	"mall/utils/logger"

	"github.com/samber/lo"
	"go.uber.org/zap"
)

func (s *Service) AddGoods(ctx context.Context, user *common.UserInfo, req *dto.AddGoodsReq) (int64, common.Errno) {
	cartId, err := s.cart.AddGoods(ctx, &do.AddGoods{
		GoodsID: req.GoodsID,
		UserID:  user.User.ID,
	})
	if err != nil {
		logger.Error("AddGoods AddGoods error", zap.Error(err), zap.Any("req", req))
		return 0, common.DatabaseErr.WithErr(err)
	}
	return cartId, common.OK
}

func (s *Service) RemoveGoods(ctx context.Context, user *common.UserInfo, req *dto.RemoveGoodsReq) common.Errno {
	err := s.cart.RemoveGoods(ctx, &do.RemoveGoods{
		ID:     req.ID,
		UserID: user.User.ID,
	})
	if err != nil {
		logger.Error("RemoveGoods RemoveGoods error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}

func (s *Service) ListGoods(ctx context.Context, user *common.UserInfo, req *dto.ListGoodsReq) (*dto.ListGoodsResp, common.Errno) {
	list, count, err := s.cart.ListGoods(ctx, &do.ListGoods{
		GoodsNameKW: req.GoodsNameKw,
		UserID:      user.User.ID,
		Pager:       req.Pager,
	})
	if err != nil {
		logger.Error("ListGoods ListGoods error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	courseIds := make([]int64, 0)
	lo.ForEach(list, func(item *model.UserCart, index int) {
		courseIds = append(courseIds, item.GoodsID)
	})
	courseList, err := s.course.GetCourseInfoByIds(ctx, courseIds)
	if err != nil {
		logger.Error("ListGoods GetCourseInfoByIds error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}

	var (
		fileKeys   = []string{}
		fileUrlMap = make(map[string]string)
	)

	lo.ForEach(courseList, func(item *model.CourseGood, index int) {
		fileKeys = append(fileKeys, item.CoverKey, item.DetailCoverKey)
	})

	fileUrlMap, err = s.storage.GetPreviewUrl(ctx, &do.GetPreviewUrl{
		Keys:        fileKeys,
		ExpireHours: 6,
	})
	if err != nil {
		logger.Error("ListGoods GetPreviewUrl error", zap.Error(err), zap.Any("fileKeys", fileKeys))
		return nil, common.ServerErr.WithErr(err)
	}
	courseMap := make(map[int64]*dto.CourseDto)
	lo.ForEach(courseList, func(item *model.CourseGood, index int) {
		var features []string
		json.Unmarshal([]byte(item.Features), &features)
		courseMap[item.ID] = &dto.CourseDto{
			ID:             item.ID,
			Name:           item.Name,
			UpdateStatus:   item.UpdateStatus,
			Features:       features,
			CoursePrice:    item.CoursePrice,
			ServiceTime:    item.ServiceTime,
			LearnTime:      item.LearnTime,
			CoverUrl:       fileUrlMap[item.CoverKey],
			DetailCoverUrl: fileUrlMap[item.DetailCoverKey],
		}
	})
	retList := make([]*dto.CartGoodsDto, 0)

	lo.ForEach(list, func(item *model.UserCart, index int) {
		retList = append(retList, &dto.CartGoodsDto{
			ID:        item.ID,
			GoodsID:   item.GoodsID,
			Quantity:  item.Quantity,
			CourseDto: courseMap[item.GoodsID],
		})
	})
	return &dto.ListGoodsResp{
		List:  retList,
		Total: count,
		Pager: req.Pager,
	}, common.OK
}
