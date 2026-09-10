package user

import (
	"context"
	"mall/adaptor/repo/model"
	"mall/common"
	"mall/consts"
	"mall/service/do"
	"mall/service/dto"
	"mall/utils/logger"
	"mall/utils/tools"

	"github.com/goccy/go-json"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

func (s *Service) GetUserByToken(ctx context.Context, token string) (*common.UserInfo, error) {
	valueStr, err := s.verify.GetUserToken(ctx, token)
	if err != nil {
		logger.Error("GetUserByToken GetUserToken", zap.Error(err), zap.String("token", token))
		return nil, common.DatabaseErr.WithErr(err)
	}
	userInfo := &common.UserInfo{}
	if err = json.Unmarshal([]byte(valueStr), userInfo); err != nil {
		logger.Error("GetUserByToken json.Unmarshal", zap.Error(err), zap.String("token", token))
		return nil, common.ServerErr.WithErr(err)
	}
	if userInfo.User.ID == 0 {
		return nil, common.AuthErr
	}
	dbUser, err := s.user.GetUserByID(ctx, userInfo.User.ID)
	if err != nil {
		logger.Error("GetUserByToken GetUserByID", zap.Error(err), zap.Int64("user_id", userInfo.User.ID))
		return nil, common.DatabaseErr.WithErr(err)
	}
	if dbUser == nil || dbUser.Status != consts.IsEnable {
		return nil, common.AuthErr
	}
	return userInfo, nil
}

func (s *Service) GetUserInfo(ctx context.Context, userID int64) (*dto.UserInfoDto, common.Errno) {
	user, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		logger.Error("GetUserInfo GetUserByID", zap.Error(err), zap.Int64("user_id", userID))
		return nil, common.DatabaseErr.WithErr(err)
	}
	userInfo, err := s.packageUserInfo(ctx, user)
	if err != nil {
		logger.Error("GetUserInfo packageUserInfo", zap.Error(err), zap.Int64("user_id", userID))
		return nil, common.DatabaseErr.WithErr(err)
	}
	return userInfo, common.OK
}

func (s *Service) CustomerUserList(ctx context.Context, req *dto.ListCustomerUserReq) (*dto.ListCustomerUserResp, common.Errno) {
	var (
		userIds      = []int64{}
		mobileSha256 string
		fileKeys     = []string{}
		fileUrlMap   = make(map[string]string)
	)
	if req.ID > 0 {
		userIds = append(userIds, req.ID)
	}
	if req.Mobile != "" {
		mobileSha256 = tools.Sha256Hash(req.Mobile)
	}
	list, count, err := s.user.ListUser(ctx, &do.ListUser{
		Pager:        req.Pager,
		UserIds:      userIds,
		NickNameKw:   req.NickNameKw,
		MobileSha256: mobileSha256,
		Status:       req.Status,
	})
	if err != nil {
		logger.Error("CustomerUserList ListUser", zap.Error(err))
		return nil, common.DatabaseErr.WithErr(err)
	}
	lo.ForEach(list, func(item *model.User, index int) {
		userIds = append(userIds, item.ID)
		fileKeys = append(fileKeys, item.IconKey)
	})
	fileUrlMap, err = s.storage.GetPreviewUrl(ctx, &do.GetPreviewUrl{
		Keys:        fileKeys,
		ExpireHours: 2,
	})
	if err != nil {
		logger.Error("CustomerUserList GetPreviewUrl", zap.Error(err))
		return nil, common.ServerErr.WithErr(err)
	}
	mobileUserList, err := s.user.GetMobileUsersByUserIds(ctx, userIds)
	if err != nil {
		logger.Error("CustomerUserList GetMobileUsersByUserIds", zap.Error(err))
		return nil, common.DatabaseErr.WithErr(err)
	}
	mobileUserMap := lo.SliceToMap(mobileUserList, func(item *model.MobileUser) (int64, *model.MobileUser) {
		return item.UserID, item
	})

	retList := make([]*dto.CustomerUserDto, 0)
	lo.ForEach(list, func(item *model.User, index int) {
		dtoUser := dto.UserDto{
			ID:          item.ID,
			NickName:    item.NickName,
			Sex:         item.Sex,
			IconUrl:     fileUrlMap[item.IconKey],
			Status:      item.Status,
			CreateAt:    item.CreateAt.UnixMilli(),
			LastLoginAt: item.LastLoginAt.UnixMilli(),
			UpdateAt:    item.UpdateAt.UnixMilli(),
			HasPassword: item.Password != "",
		}
		customerUser := &dto.CustomerUserDto{
			UserDto: dtoUser,
		}
		mobileUser, ok := mobileUserMap[item.ID]
		if ok {
			mobile, err := tools.AESDecrypt(mobileUser.MobileAes, []byte(s.conf.BizConf.MobileSecret))
			if err != nil {
				logger.Error("CustomerUserList AESDecrypt", zap.Error(err), zap.String("mobile_aes", mobileUser.MobileAes))
			} else {
				customerUser.Mobile = string(mobile)
			}
		}
		retList = append(retList, customerUser)
	})
	return &dto.ListCustomerUserResp{
		List:  retList,
		Pager: req.Pager,
		Total: count,
	}, common.OK
}
