package goods

import (
	"context"
	"encoding/json"

	"mall/adaptor"
	"mall/adaptor/repo/model"
	"mall/adaptor/repo/query"
	"mall/consts"
	"mall/service/do"
	"mall/utils/tools"
	"time"

	"github.com/gogf/gf/util/gconv"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ICourse interface {
	CreateCourse(ctx context.Context, req *do.CreateCourse) (int64, error)
	GetCourseInfoById(ctx context.Context, id int64) (*model.CourseGood, error)
	GetCourseInfoByIdAndStatus(ctx context.Context, id int64, status int32) (*model.CourseGood, error)
	GetCourseInfoByIds(ctx context.Context, ids []int64) ([]*model.CourseGood, error)
	UpdateCourse(ctx context.Context, req *do.UpdateCourse) error
	UpdateCourseStatus(ctx context.Context, req *do.UpdateCourseStatus) error
	ListCourse(ctx context.Context, req *do.CourseList) ([]*model.CourseGood, int64, error)

	AddCatalog(ctx context.Context, req *do.AddCatalog) (int64, error)
	UpdateCatalog(ctx context.Context, req *do.UpdateCatalog) error
	DeleteCatalog(ctx context.Context, req *do.DeleteCatalog) error
	UpdateCatalogSort(ctx context.Context, req *do.UpdateCatalogSort) error
	AddCatalogLesson(ctx context.Context, req *do.AddCatalogLesson) error
	UpdateCatalogLesson(ctx context.Context, req *do.UpdateCatalogLesson) error
	RemoveCatalogLesson(ctx context.Context, req *do.RemoveCatalogLesson) error

	GetCourseLessons(ctx context.Context, courseId int64) ([]*model.CourseLesson, error)
	GetCatalogsByCourseId(ctx context.Context, courseId int64) ([]*model.CourseCatalog, error)
}

type Course struct {
	db *gorm.DB
}

func NewCourse(adaptor adaptor.IAdaptor) *Course {
	return &Course{
		db: adaptor.GetDB(),
	}
}

func (s *Course) GetCourseInfoById(ctx context.Context, id int64) (*model.CourseGood, error) {
	qs := query.Use(s.db).CourseGood
	return qs.WithContext(ctx).Where(qs.ID.Eq(id)).First()
}

func (s *Course) GetCourseInfoByIdAndStatus(ctx context.Context, id int64, status int32) (*model.CourseGood, error) {
	qs := query.Use(s.db).CourseGood
	return qs.WithContext(ctx).Where(qs.ID.Eq(id), qs.Status.Eq(status)).First()
}

func (s *Course) GetCourseInfoByIds(ctx context.Context, ids []int64) ([]*model.CourseGood, error) {
	qs := query.Use(s.db).CourseGood
	return qs.WithContext(ctx).Where(qs.ID.In(ids...)).Find()
}

func (s *Course) CreateCourse(ctx context.Context, req *do.CreateCourse) (int64, error) {
	timeNow := time.Now()
	files, err := json.Marshal(req.Files)
	if err != nil {
		return 0, err
	}
	addCourse := &model.CourseGood{
		Name:           req.Name,
		CoverKey:       req.CoverKey,
		DetailCoverKey: req.DetailCoverKey,
		Detail:         req.Detail,
		CoursePrice:    req.CoursePrice,
		ServiceTime:    req.ServiceTime,
		LearnTime:      req.LearnTime,
		Status:         consts.IsDisable,
		UpdateStatus:   req.UpdateStatus,
		Features:       gconv.String(req.Features),
		Files:          string(files),
		CreateAt:       timeNow,
		CreateBy:       req.UserID,
		UpdateAt:       timeNow,
		UpdateBy:       req.UserID,
	}
	err = s.db.WithContext(ctx).Create(addCourse).Error
	return addCourse.ID, err
}

func (s *Course) UpdateCourseStatus(ctx context.Context, req *do.UpdateCourseStatus) error {
	timeNow := time.Now()
	qs := query.Use(s.db).CourseGood
	updateMap := map[string]interface{}{
		qs.UpdateBy.ColumnName().String(): req.UserID,
		qs.UpdateAt.ColumnName().String(): timeNow,
		qs.Status.ColumnName().String():   req.Status,
	}
	_, err := qs.WithContext(ctx).Where(qs.ID.Eq(req.ID)).Updates(updateMap)
	return err
}

func (s *Course) UpdateCourse(ctx context.Context, req *do.UpdateCourse) error {
	timeNow := time.Now()
	files, err := json.Marshal(req.Files)
	if err != nil {
		return err
	}
	qs := query.Use(s.db).CourseGood
	updateMap := map[string]interface{}{
		qs.UpdateBy.ColumnName().String(): req.UserID,
		qs.UpdateAt.ColumnName().String(): timeNow,
	}
	if req.Name != "" {
		updateMap[qs.Name.ColumnName().String()] = req.Name
	}
	if req.CoverKey != "" {
		updateMap[qs.CoverKey.ColumnName().String()] = req.CoverKey
	}
	if req.DetailCoverKey != "" {
		updateMap[qs.DetailCoverKey.ColumnName().String()] = req.DetailCoverKey
	}
	if req.Detail != "" {
		updateMap[qs.Detail.ColumnName().String()] = req.Detail
	}
	if req.CoursePrice != 0 {
		updateMap[qs.CoursePrice.ColumnName().String()] = req.CoursePrice
	}
	if req.ServiceTime != 0 {
		updateMap[qs.ServiceTime.ColumnName().String()] = req.ServiceTime
	}
	if req.LearnTime != 0 {
		updateMap[qs.LearnTime.ColumnName().String()] = req.LearnTime
	}
	if req.Sort != 0 {
		updateMap[qs.Sort.ColumnName().String()] = req.Sort
	}
	if req.UpdateStatus != 0 {
		updateMap[qs.UpdateStatus.ColumnName().String()] = req.UpdateStatus
	}
	if len(req.Features) != 0 {
		updateMap[qs.Features.ColumnName().String()] = gconv.String(req.Features)
	}
	updateMap[qs.Files.ColumnName().String()] = string(files)
	_, err = qs.WithContext(ctx).Where(qs.ID.Eq(req.ID)).Updates(updateMap)
	return err
}

func (s *Course) ListCourse(ctx context.Context, req *do.CourseList) ([]*model.CourseGood, int64, error) {
	qs := query.Use(s.db).CourseGood
	uqs := query.Use(s.db).UserCourseGood
	tx := s.db.WithContext(ctx).Model(&model.CourseGood{})
	if req.ID != 0 {
		tx = tx.Where(qs.ID.Eq(req.ID))
	}
	if req.NameKw != "" {
		tx = tx.Where(qs.Name.Like(tools.GetAllLike(req.NameKw)))
	}
	if req.StartCreateTime != 0 && req.EndCreateTime != 0 {
		tx = tx.Where(qs.CreateAt.Between(time.UnixMilli(req.StartCreateTime), time.UnixMilli(req.EndCreateTime)))
	}
	if req.StartUpdateTime != 0 && req.EndUpdateTime != 0 {
		tx = tx.Where(qs.UpdateAt.Between(time.UnixMilli(req.StartUpdateTime), time.UnixMilli(req.EndUpdateTime)))
	}
	if req.UpdateStatus != 0 {
		tx = tx.Where(qs.UpdateStatus.Eq(req.UpdateStatus))
	}
	if req.Status != 0 {
		tx = tx.Where(qs.Status.Eq(req.Status))
	}
	if req.UserID != 0 && req.ExcludePurchasedGoods {
		existSubQuery := s.db.Model(&model.UserCourseGood{}).Select("1").Where(
			uqs.UserID.Eq(req.UserID),
			uqs.GoodsType.Eq(consts.CourseGood),
			uqs.GoodsID.EqCol(qs.ID),
		)
		tx = tx.Where("NOT EXISTS (?)", existSubQuery)
	}
	var count int64
	err := tx.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	tx = tx.Order(
		clause.OrderBy{Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: qs.Status.ColumnName().String()}, Desc: true},
			{Column: clause.Column{Name: qs.Sort.ColumnName().String()}, Desc: false},
			{Column: clause.Column{Name: qs.ID.ColumnName().String()}, Desc: true},
		}})
	retList := make([]*model.CourseGood, 0)
	err = tx.Offset(req.GetOffset()).Limit(req.Limit).Find(&retList).Error
	return retList, count, err
}

func (s *Course) AddCatalog(ctx context.Context, req *do.AddCatalog) (int64, error) {
	timeNow := time.Now()
	var maxSort int32
	err := s.db.WithContext(ctx).Model(&model.CourseCatalog{}).
		Where("course_id = ? AND parent_id = ?", req.CourseID, req.ParentID).
		Select("COALESCE(MAX(sort), 0)").
		Scan(&maxSort).Error
	if err != nil {
		return 0, err
	}
	addCatalog := &model.CourseCatalog{
		CourseID: req.CourseID,
		Name:     req.Name,
		Level:    req.Level,
		ParentID: req.ParentID,
		Sort:     maxSort + 1,
		UpdateAt: timeNow,
		UpdateBy: req.UserID,
	}
	err = s.db.WithContext(ctx).Create(addCatalog).Error
	return addCatalog.ID, err
}

func (s *Course) UpdateCatalog(ctx context.Context, req *do.UpdateCatalog) error {
	timeNow := time.Now()
	qs := query.Use(s.db).CourseCatalog
	updateMap := map[string]interface{}{
		qs.UpdateBy.ColumnName().String(): req.UserID,
		qs.UpdateAt.ColumnName().String(): timeNow,
	}
	if req.Name != "" {
		updateMap[qs.Name.ColumnName().String()] = req.Name
	}

	_, err := qs.WithContext(ctx).Where(qs.ID.Eq(req.ID)).Updates(updateMap)
	return err
}

func (s *Course) DeleteCatalog(ctx context.Context, req *do.DeleteCatalog) error {
	qs := query.Use(s.db).CourseCatalog
	_, err := qs.WithContext(ctx).Where(qs.ID.Eq(req.ID)).Delete()
	return err
}

func (s *Course) GetCourseLessons(ctx context.Context, courseId int64) ([]*model.CourseLesson, error) {
	qs := query.Use(s.db).CourseLesson
	return qs.WithContext(ctx).Where(qs.CourseGoodsID.Eq(courseId)).Order(qs.Sort).Find()
}

func (s *Course) GetCatalogsByCourseId(ctx context.Context, courseId int64) ([]*model.CourseCatalog, error) {
	qs := query.Use(s.db).CourseCatalog
	return qs.WithContext(ctx).Where(qs.CourseID.Eq(courseId)).Order(qs.Sort, qs.ID).Find()
}

func (s *Course) UpdateCatalogSort(ctx context.Context, req *do.UpdateCatalogSort) error {
	timeNow := time.Now()
	cqs := query.Use(s.db).CourseCatalog
	lqs := query.Use(s.db).CourseLesson
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, catalog := range req.SortList {
			err := tx.Model(&model.CourseCatalog{}).Where(cqs.ID.Eq(catalog.Id)).Updates(map[string]interface{}{
				cqs.Sort.ColumnName().String():     catalog.Sort,
				cqs.ParentID.ColumnName().String(): catalog.ParentId,
				cqs.Level.ColumnName().String():    catalog.Level,
				cqs.UpdateAt.ColumnName().String(): timeNow,
				cqs.UpdateBy.ColumnName().String(): req.UserID,
			}).Error
			if err != nil {
				return err
			}
			for _, clesson := range catalog.Lessons {
				err = tx.Model(&model.CourseLesson{}).Where(lqs.ID.Eq(clesson.ID)).Updates(map[string]interface{}{
					lqs.Sort.ColumnName().String():     clesson.Sort,
					lqs.UpdateAt.ColumnName().String(): timeNow,
					lqs.UpdateBy.ColumnName().String(): req.UserID,
				}).Error
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *Course) AddCatalogLesson(ctx context.Context, req *do.AddCatalogLesson) error {
	timeNow := time.Now()
	qs := query.Use(s.db).CourseLesson
	var maxSort int32
	err := s.db.WithContext(ctx).Model(&model.CourseLesson{}).
		Where("catalog_id = ?", req.CatalogID).
		Select("COALESCE(MAX(sort), 0)").
		Scan(&maxSort).Error
	if err != nil {
		return err
	}
	addList := make([]*model.CourseLesson, 0, len(req.LessonMap))
	for index, lessonID := range req.LessonIDs {
		name := req.LessonMap[lessonID]
		addList = append(addList, &model.CourseLesson{
			CatalogID:     req.CatalogID,
			CourseGoodsID: req.CourseID,
			EnableTrial:   0,
			Name:          name,
			LessonID:      lessonID,
			ShowTime:      timeNow.Add(time.Minute * 5),
			Sort:          maxSort + int32(index) + 1,
			UpdateAt:      timeNow,
			UpdateBy:      req.UserID,
		})
	}
	return qs.WithContext(ctx).CreateInBatches(addList, 100)
}

func (s *Course) UpdateCatalogLesson(ctx context.Context, req *do.UpdateCatalogLesson) error {
	timeNow := time.Now()
	qs := query.Use(s.db).CourseLesson
	updateMap := map[string]interface{}{
		qs.UpdateBy.ColumnName().String(): req.UserID,
		qs.UpdateAt.ColumnName().String(): timeNow,
	}
	if req.EnableTrial != 0 {
		updateMap[qs.EnableTrial.ColumnName().String()] = req.EnableTrial
	}
	if req.Name != "" {
		updateMap[qs.Name.ColumnName().String()] = req.Name
	}
	if req.ShowTime != 0 {
		updateMap[qs.ShowTime.ColumnName().String()] = time.UnixMilli(req.ShowTime)
	}
	_, err := qs.WithContext(ctx).Where(qs.ID.Eq(req.ID)).Updates(updateMap)
	return err
}
func (s *Course) RemoveCatalogLesson(ctx context.Context, req *do.RemoveCatalogLesson) error {
	qs := query.Use(s.db).CourseLesson
	_, err := qs.WithContext(ctx).Where(qs.ID.In(req.IDs...)).Delete()
	if err != nil {
		return err
	}
	return nil
}
