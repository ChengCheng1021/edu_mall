package goods

import (
	"context"
	"mall/adaptor/repo/model"
	"mall/common"
	"mall/consts"
	"mall/service/do"
	"mall/service/dto"
	"mall/utils/logger"
	"time"

	"go.uber.org/zap"
)

func (s *Service) GetLessonLearnInfo(ctx context.Context, user *common.UserInfo, req *dto.LessonLearnInfoReq) (*dto.LessonLearnInfoResp, common.Errno) {
	if user == nil || user.User.ID <= 0 {
		return nil, common.AuthErr
	}
	if req.CourseID <= 0 || req.LessonID <= 0 {
		return nil, common.ParamErr.WithMsg("course_id or lesson_id invalid")
	}
	_, errno := s.getLessonLearnSessionInfo(ctx, user.User.ID, req.CourseID, req.LessonID)
	if errno.NotOk() {
		return nil, errno
	}
	session, err := s.rdsLearn.GetLessonLearnSession(ctx, user.User.ID, req.CourseID, req.LessonID)
	if err != nil {
		logger.Error("GetLessonLearnInfo GetLessonLearnSession error", zap.Error(err), zap.Any("req", req), zap.Int64("user_id", user.User.ID))
		return nil, common.RedisErr.WithErr(err)
	}
	progress, err := s.learn.GetLessonLearnProgress(ctx, user.User.ID, req.CourseID, req.LessonID)
	if err != nil {
		logger.Error("GetLessonLearnInfo GetLessonLearnProgress error", zap.Error(err), zap.Any("req", req), zap.Int64("user_id", user.User.ID))
		return nil, common.DatabaseErr.WithErr(err)
	}
	resp := &dto.LessonLearnInfoResp{
		CourseID:    req.CourseID,
		LessonID:    req.LessonID,
		LearnStatus: consts.LessonLearnStatusLearning,
	}
	if progress != nil {
		resp.PlayPosition = progress.PlayPosition
		resp.LearnStatus = progress.LearnStatus
	}
	if session != nil {
		resp.PlayPosition = session.PlayPosition
		resp.LastType = session.LastType
		resp.EntryTime = session.EntryTime
		resp.LastReportTime = session.LastEventTime
		resp.InLearning = true
	}
	return resp, common.OK
}

func (s *Service) getLessonLearnSessionInfo(ctx context.Context, userID, courseID, lessonID int64) (*do.LessonLearnSessionInfo, common.Errno) {
	hasCourse, err := s.userCourse.HasValidUserCourse(ctx, userID, courseID)
	if err != nil {
		logger.Error("getLessonLearnSessionInfo HasValidUserCourse error", zap.Error(err), zap.Int64("user_id", userID), zap.Int64("course_id", courseID))
		return nil, common.DatabaseErr.WithErr(err)
	}
	courseLessons, err := s.course.GetCourseLessons(ctx, courseID)
	if err != nil {
		logger.Error("getLessonLearnSessionInfo GetCourseLessons error", zap.Error(err), zap.Int64("course_id", courseID))
		return nil, common.DatabaseErr.WithErr(err)
	}
	var currentLesson *model.CourseLesson
	for _, item := range courseLessons {
		if item == nil {
			continue
		}
		if item.LessonID == lessonID {
			currentLesson = item
		}
	}
	if currentLesson == nil {
		return nil, common.ParamErr.WithMsg("lesson_id invalid")
	}
	if !hasCourse && currentLesson.EnableTrial != consts.IsEnable {
		return nil, common.PermissionErr.WithMsg("course not purchased or expired")
	}
	if currentLesson.ShowTime.UnixMilli() > time.Now().UnixMilli() {
		return nil, common.PermissionErr.WithMsg("lesson not available yet")
	}
	lessonMap, err := s.lesson.GetLessonByIds(ctx, []int64{lessonID})
	if err != nil {
		logger.Error("getLessonLearnSessionInfo GetLessonByIds error", zap.Error(err), zap.Int64("lesson_id", lessonID))
		return nil, common.DatabaseErr.WithErr(err)
	}
	lesson, ok := lessonMap[lessonID]
	if !ok || lesson == nil {
		return nil, common.ParamErr.WithMsg("lesson_id invalid")
	}
	catalogList, err := s.course.GetCatalogsByCourseId(ctx, courseID)
	if err != nil {
		logger.Error("getLessonLearnSessionInfo GetCatalogsByCourseId error", zap.Error(err), zap.Int64("course_id", courseID))
		return nil, common.DatabaseErr.WithErr(err)
	}
	catalogMap := make(map[int64]*model.CourseCatalog, len(catalogList))
	for _, item := range catalogList {
		if item == nil {
			continue
		}
		catalogMap[item.ID] = item
	}
	secondCategoryID := currentLesson.CatalogID
	firstCategoryID := currentLesson.CatalogID
	if catalog, ok := catalogMap[currentLesson.CatalogID]; ok && catalog != nil {
		if catalog.Level == 2 {
			secondCategoryID = catalog.ID
			firstCategoryID = catalog.ParentID
		} else {
			firstCategoryID = catalog.ID
			secondCategoryID = 0
		}
	}
	if firstCategoryID <= 0 {
		firstCategoryID = secondCategoryID
	}
	return &do.LessonLearnSessionInfo{
		CourseID:         courseID,
		LessonID:         lessonID,
		FirstCategoryID:  firstCategoryID,
		SecondCategoryID: secondCategoryID,
		LessonDuration:   int64(lesson.Duration),
	}, common.OK
}
