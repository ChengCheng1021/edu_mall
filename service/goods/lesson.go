package goods

import (
	"context"
	"encoding/json"
	"errors"
	"mall/adaptor/repo/model"
	"mall/common"
	"mall/service/do"
	"mall/service/dto"
	"mall/utils/logger"
	"mall/utils/pool"
	"mall/utils/tools"

	"github.com/samber/lo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *Service) CategoryCreate(ctx context.Context, req *dto.AddCategoryReq) (int64, common.Errno) {
	categoryID, err := s.lesson.AddCategory(ctx, &do.AddCategory{
		Name:     req.Name,
		Level:    req.Level,
		ParentID: req.ParentID,
		Sort:     req.Sort,
	})
	if err != nil {
		logger.Error("CategoryCreate AddCategory error", zap.Error(err), zap.Any("req", req))
		return 0, common.DatabaseErr.WithErr(err)
	}
	return categoryID, common.OK
}
func (s *Service) CategoryUpdate(ctx context.Context, req *dto.UpdateCategoryReq) common.Errno {
	err := s.lesson.UpdateCategory(ctx, &do.UpdateCategory{
		ID:   req.ID,
		Name: req.Name,
	})
	if err != nil {
		logger.Error("CategoryUpdate UpdateCategory error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}
func (s *Service) CategoryDelete(ctx context.Context, req *dto.DeleteCategoryReq) common.Errno {
	// 如果删除的是一级目录 并且 下面还有数据那就全部删除
	allIDs := make([]int64, 0, len(req.IDs))
	allIDs = append(allIDs, req.IDs...)
	parentIDs := req.IDs

	for len(parentIDs) > 0 {
		children, err := s.lesson.FindByParentIDs(ctx, parentIDs)
		if err != nil {
			logger.Error("FindByParentIDs FindByParentIDs error", zap.Error(err), zap.Any("req", req))
			return common.DatabaseErr.WithErr(err)
		}

		if len(children) == 0 {
			break
		}

		childIDs := make([]int64, 0, len(children))
		for _, child := range children {
			childIDs = append(childIDs, child.ID)
		}

		allIDs = append(allIDs, childIDs...)

		// 下一轮：当前子节点变成下一层的父节点
		parentIDs = childIDs
	}
	err := s.lesson.DeleteCategory(ctx, &do.DeleteCategory{IDs: allIDs})
	if err != nil {
		logger.Error("CategoryDelete DeleteCategory error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}

func (s *Service) CategorySort(ctx context.Context, sortList []dto.UpdateSort) common.Errno {
	updateSorts := make([]*do.UpdateSort, 0)
	lo.ForEach(sortList, func(item dto.UpdateSort, index int) {
		updateSorts = append(updateSorts, &do.UpdateSort{
			ID:   item.ID,
			Sort: item.Sort,
		})
	})
	err := s.lesson.UpdateCategorySort(ctx, updateSorts)
	if err != nil {
		logger.Error("CategorySort UpdateCategorySort error", zap.Error(err), zap.Any("sortList", sortList))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}
func (s *Service) CategoryList(ctx context.Context, req *dto.ListCategoryReq) ([]*dto.CategoryDto, common.Errno) {
	list, err := s.lesson.ListCategory(ctx, &do.ListCategory{Pager: req.Pager})
	if err != nil {
		logger.Error("CategoryList ListCategory error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}
	return lo.Map(list, func(item *model.LessonCategory, index int) *dto.CategoryDto {
		return &dto.CategoryDto{
			ID:       item.ID,
			Name:     item.Name,
			Sort:     item.Sort,
			ParentID: item.ParentID,
			Level:    item.Level,
		}
	}), common.OK
}

func (s *Service) CreateLesson(ctx context.Context, user *common.AdminUser, req *dto.CreateLessonReq) (int64, common.Errno) {
	lessonID, err := s.lesson.CreateLesson(ctx, &do.CreateLesson{
		Name:          req.Name,
		Detail:        req.Detail,
		CategoryID:    req.CategoryID,
		VideoKey:      req.VideoKey,
		Duration:      req.Duration,
		VideoFileName: req.VideoFileName,
		UserID:        user.UserID,
		Attachments: lo.Map(req.Attachments, func(item dto.Attachment, index int) do.Attachment {
			return do.Attachment{
				OriginName: item.OriginName,
				FileKey:    item.FileKey,
			}
		}),
		Chapters: lo.Map(req.Chapters, func(item dto.LessonChapter, index int) do.LessonChapter {
			return do.LessonChapter{
				ID:            tools.UUIDHex(),
				Name:          item.Name,
				BeginPosition: item.BeginPosition,
				EndPosition:   item.EndPosition,
			}
		}),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return 0, common.ParamErr.WithMsg("name duplicated")
		}
		logger.Error("CreateLesson CreateLesson error", zap.Error(err), zap.Any("req", req))
		return 0, common.DatabaseErr.WithErr(err)
	}
	return lessonID, common.OK
}

func (s *Service) LessonList(ctx context.Context, req *dto.ListLessonReq) (*dto.ListLessonResp, common.Errno) {
	var (
		categoryIDs = make([]int64, 0)
		excludeIds  = make([]int64, 0)
		userIds     = make([]int64, 0)
		fileKeys    = make([]string, 0)
		cateNameMap = make(map[int64]string)
		userNameMap = make(map[int64]string)
		fileNameMap = make(map[string]string)
	)
	if req.CategoryID != 0 {
		categoryIDs = append(categoryIDs, req.CategoryID)
		if req.OnView {
			tempIds, err := s.lesson.GetChildCategoryIds(ctx, []int64{req.CategoryID})
			if err != nil {
				logger.Error("LessonList GetChildCategoryIds error", zap.Error(err), zap.Any("req", req))
				return nil, common.DatabaseErr.WithErr(err)
			}
			categoryIDs = append(categoryIDs, tempIds...)
		}
	}
	if req.CourseID > 0 {
		courseLessons, err := s.course.GetCourseLessons(ctx, req.CourseID)
		if err != nil {
			logger.Error("LessonList GetCourseLessons error", zap.Error(err), zap.Any("req", req))
			return nil, common.DatabaseErr.WithErr(err)
		}
		excludeIds = lo.Map(courseLessons, func(item *model.CourseLesson, index int) int64 {
			return item.LessonID
		})
	}
	list, count, err := s.lesson.LessonList(ctx, &do.ListLesson{
		ExcludeIds:      excludeIds,
		Pager:           req.Pager,
		CategoryIDs:     categoryIDs,
		Status:          req.Status,
		ID:              req.ID,
		NameKw:          req.NameKw,
		StartCreateTime: req.StartCreateTime,
		EndCreateTime:   req.EndCreateTime,
		StartUpdateTime: req.StartUpdateTime,
		EndUpdateTime:   req.EndUpdateTime,
	})
	if err != nil {
		logger.Error("LessonList ListLesson error", zap.Error(err), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(err)
	}

	lo.ForEach(list, func(item *model.Lesson, index int) {
		categoryIDs = append(categoryIDs, item.CategoryID)
		userIds = append(userIds, item.CreateBy, item.UpdateBy)
		fileKeys = append(fileKeys, item.VideoKey)

		attachments := make([]dto.Attachment, 0)
		json.Unmarshal([]byte(item.Attachments), &attachments)
		lo.ForEach(attachments, func(item dto.Attachment, index int) {
			fileKeys = append(fileKeys, item.FileKey)
		})
	})
	categoryIDs = lo.Uniq(categoryIDs)
	userIds = lo.Uniq(userIds)

	tempPool := pool.NewPoolWithSize(3)
	defer tempPool.Release()
	tempPool.RunGo(func() {
		temp, err := s.lesson.GetCategoryNameMap(ctx, categoryIDs)
		if err != nil {
			logger.Error("LessonList GetCategoryNameMap error", zap.Error(err), zap.Any("categoryIDs", categoryIDs))
			return
		}
		cateNameMap = temp
	})
	tempPool.RunGo(func() {
		temp, err := s.adminUser.GetUserNameMap(ctx, userIds)
		if err != nil {
			logger.Error("LessonList GetUserNameMap error", zap.Error(err), zap.Any("userIds", userIds))
			return
		}
		userNameMap = temp
	})
	tempPool.RunGo(func() {
		temp, err := s.storage.GetPreviewUrl(ctx, &do.GetPreviewUrl{
			Keys:        fileKeys,
			ExpireHours: 6,
		})
		if err != nil {
			logger.Error("LessonList GetFileNameMap error", zap.Error(err), zap.Any("fileKeys", fileKeys))
			return
		}
		fileNameMap = temp
	})
	tempPool.Wait()

	retList := make([]*dto.LessonDto, 0)
	lo.ForEach(list, func(item *model.Lesson, index int) {
		attachments := make([]dto.Attachment, 0)
		chapters := make([]dto.LessonChapter, 0)
		json.Unmarshal([]byte(item.Attachments), &attachments)
		json.Unmarshal([]byte(item.Chapters), &chapters)
		lo.ForEach(attachments, func(item dto.Attachment, index int) {
			attachments[index].FileUrl = fileNameMap[item.FileKey]
		})
		retList = append(retList, &dto.LessonDto{
			ID:            item.ID,
			Name:          item.Name,
			Detail:        item.Detail,
			Duration:      item.Duration,
			VideoKey:      item.VideoKey,
			VideoFileName: item.VideoFileName,
			VideoUrl:      fileNameMap[item.VideoKey],
			Status:        item.Status,
			CategoryID:    item.CategoryID,
			CategoryName:  cateNameMap[item.CategoryID],
			CreateBy:      item.CreateBy,
			UpdateBy:      item.UpdateBy,
			Attachments:   attachments,
			Chapters:      chapters,
			CreateUpdateName: common.CreateUpdateName{
				CreateName: userNameMap[item.CreateBy],
				UpdateName: userNameMap[item.UpdateBy],
			},
			CreateAt: item.CreateAt.UnixMilli(),
			UpdateAt: item.UpdateAt.UnixMilli(),
		})
	})
	return &dto.ListLessonResp{
		List:  retList,
		Total: count,
	}, common.OK
}

func (s *Service) LessonInfo(ctx context.Context, req *dto.LessonInfoReq) (*dto.LessonDto, common.Errno) {
	lessonID := lo.Ternary(req.LessonID > 0, req.LessonID, req.ID)
	resp, errno := s.LessonList(ctx, &dto.ListLessonReq{
		ID: lessonID,
	})
	if errno.NotOk() {
		logger.Error("LessonInfo LessonList error", zap.Error(errno), zap.Any("req", req))
		return nil, common.DatabaseErr.WithErr(errno)
	}
	if len(resp.List) == 0 {
		return nil, common.ParamErr.WithMsg("id invalid")
	}
	return resp.List[0], common.OK
}

func (s *Service) LessonDetail(ctx context.Context, req *dto.LessonInfoReq) (any, bool, common.Errno) {
	lessonID := lo.Ternary(req.LessonID > 0, req.LessonID, req.ID)
	resp, errno := s.LessonList(ctx, &dto.ListLessonReq{
		ID: lessonID,
	})
	if errno.NotOk() {
		logger.Error("LessonInfo LessonList error", zap.Error(errno), zap.Any("req", req))
		return nil, true, common.DatabaseErr.WithErr(errno)
	}
	if len(resp.List) == 0 {
		return nil, true, common.ParamErr.WithMsg("id invalid")
	}

	lessonDto := resp.List[0]

	// 加密返回
	//str := gconv.String(resp.List[0])
	//objAes, err := tools.AESEncrypt(str, []byte(s.conf.BizConf.BizSecret))
	//if err != nil {
	//	logger.Error("LessonInfo AESEncrypt error", zap.Error(err), zap.Any("req", req))
	//	return nil, true, common.ServerErr.WithErr(err)
	//}
	return lessonDto, true, common.OK
}

func (s *Service) MoveLesson(ctx context.Context, user *common.AdminUser, req *dto.MoveLessonReq) common.Errno {
	category, err := s.lesson.GetCategoryById(ctx, req.CategoryID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("MoveLesson GetCategoryNameMap error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	if category == nil {
		return common.ParamErr.WithMsg("category_id invalid")
	}

	err = s.lesson.MoveLesson(ctx, &do.MoveLesson{
		UserID:     user.UserID,
		LessonIds:  req.LessonIds,
		CategoryID: req.CategoryID,
	})
	if err != nil {
		logger.Error("MoveLesson MoveLesson error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}

func (s *Service) UpdateLessonStatus(ctx context.Context, user *common.AdminUser, req *dto.UpdateLessonStatusReq) common.Errno {
	err := s.lesson.UpdateLessonStatus(ctx, &do.UpdateLessonStatus{
		UserID: user.UserID,
		ID:     req.ID,
		Status: req.Status,
	})
	if err != nil {
		logger.Error("UpdateLessonStatus UpdateLessonStatus error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	return common.OK
}

func (s *Service) UpdateLesson(ctx context.Context, user *common.AdminUser, req *dto.UpdateLessonReq) common.Errno {
	// 获取录播信息 方便下面对比 新旧视频是否一致
	lesson, err := s.lesson.GetLessonById(ctx, req.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("UpdateLesson GetLessonById error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	if lesson == nil {
		return common.ParamErr.WithMsg("id invalid")
	}
	err = s.lesson.UpdateLesson(ctx, &do.UpdateLesson{
		Attachments: lo.Map(req.Attachments, func(item dto.Attachment, index int) do.Attachment {
			return do.Attachment{
				OriginName: item.OriginName,
				FileKey:    item.FileKey,
			}
		}),
		Chapters: lo.Map(req.Chapters, func(item dto.LessonChapter, index int) do.LessonChapter {
			return do.LessonChapter{
				ID:            lo.Ternary(item.ID == "", tools.UUIDHex(), item.ID),
				Name:          item.Name,
				BeginPosition: item.BeginPosition,
				EndPosition:   item.EndPosition,
			}
		}),
		UserID:   user.UserID,
		ID:       req.ID,
		Name:     req.Name,
		Detail:   req.Detail,
		VideoKey: req.VideoKey,
		Duration: req.Duration,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return common.ParamErr.WithMsg("name d -a" +
				"uplicated")
		}
		logger.Error("UpdateLesson UpdateLesson error", zap.Error(err), zap.Any("req", req))
		return common.DatabaseErr.WithErr(err)
	}
	if lesson.VideoKey != req.VideoKey { // 视频已经被更换，删除cos中的旧视频文件
		err = s.storage.DeleteFile(ctx, &do.DeleteFile{
			Keys: []string{lesson.VideoKey},
		})
		if err != nil {
			logger.Error("UpdateLesson DeleteFile error", zap.Error(err), zap.Any("req", req))
			return common.OK
		}
		err = s.upload.DeleteUploadFile(ctx, []string{lesson.VideoKey})
		if err != nil {
			logger.Error("UpdateLesson DeleteUploadFile error", zap.Error(err), zap.Any("req", req))
			return common.OK
		}
	}
	return common.OK
}
