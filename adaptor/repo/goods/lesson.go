package goods

import (
	"context"
	"github.com/gogf/gf/util/gconv"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"mall/adaptor"
	"mall/adaptor/repo/model"
	"mall/adaptor/repo/query"
	"mall/consts"
	"mall/service/do"
	"mall/utils/tools"
	"time"
)

type ILesson interface {
	AddCategory(ctx context.Context, req *do.AddCategory) (int64, error)
	UpdateCategory(ctx context.Context, req *do.UpdateCategory) error
	DeleteCategory(ctx context.Context, req *do.DeleteCategory) error
	ListCategory(ctx context.Context, req *do.ListCategory) ([]*model.LessonCategory, error)
	UpdateCategorySort(ctx context.Context, sortList []*do.UpdateSort) error
	GetChildCategoryIds(ctx context.Context, parentIDs []int64) ([]int64, error)
	GetCategoryNameMap(ctx context.Context, ids []int64) (map[int64]string, error)
	GetCategoryById(ctx context.Context, id int64) (*model.LessonCategory, error)

	MoveLesson(ctx context.Context, req *do.MoveLesson) error
	GetLessonById(ctx context.Context, id int64) (*model.Lesson, error)
	GetLessonByIds(ctx context.Context, ids []int64) (map[int64]*model.Lesson, error)
	UpdateLesson(ctx context.Context, req *do.UpdateLesson) error
	UpdateLessonStatus(ctx context.Context, req *do.UpdateLessonStatus) error
	CreateLesson(ctx context.Context, req *do.CreateLesson) (int64, error)
	LessonList(ctx context.Context, req *do.ListLesson) ([]*model.Lesson, int64, error)
}

type Lesson struct {
	db *gorm.DB
}

func NewLesson(adaptor adaptor.IAdaptor) *Lesson {
	return &Lesson{
		db: adaptor.GetDB(),
	}
}

func (l *Lesson) AddCategory(ctx context.Context, req *do.AddCategory) (int64, error) {
	addObj := &model.LessonCategory{
		Name:     req.Name,
		ParentID: lo.Ternary(req.ParentID == 0, -1, req.ParentID),
		Level:    req.Level,
		Sort:     req.Sort,
	}
	err := l.db.WithContext(ctx).Create(addObj).Error
	return addObj.ID, err
}

func (l *Lesson) UpdateCategory(ctx context.Context, req *do.UpdateCategory) error {
	qs := query.Use(l.db).LessonCategory
	_, err := qs.WithContext(ctx).Where(qs.ID.Eq(req.ID)).Update(qs.Name, req.Name)
	return err
}

func (l *Lesson) DeleteCategory(ctx context.Context, req *do.DeleteCategory) error {
	qs := query.Use(l.db).LessonCategory
	_, err := qs.WithContext(ctx).Where(qs.ID.In(req.IDs...)).Delete()
	return err
}

func (l *Lesson) ListCategory(ctx context.Context, req *do.ListCategory) ([]*model.LessonCategory, error) {
	qs := query.Use(l.db).LessonCategory
	list, err := qs.WithContext(ctx).Order(qs.Sort).Find()
	return list, err
}

func (l *Lesson) GetCategoryNameMap(ctx context.Context, ids []int64) (map[int64]string, error) {
	qs := query.Use(l.db).LessonCategory
	list, err := qs.WithContext(ctx).Select(qs.ID, qs.Name).Where(qs.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	return lo.SliceToMap(list, func(item *model.LessonCategory) (int64, string) {
		return item.ID, item.Name
	}), err
}

func (l *Lesson) GetCategoryById(ctx context.Context, id int64) (*model.LessonCategory, error) {
	qs := query.Use(l.db).LessonCategory
	return qs.WithContext(ctx).Where(qs.ID.Eq(id)).First()
}

func (l *Lesson) GetChildCategoryIds(ctx context.Context, parentIDs []int64) ([]int64, error) {
	var descendants []*model.LessonCategory
	rawSQL := `
    WITH RECURSIVE cte AS (
      SELECT
        id,
        parent_id,
        name,
        sort
      FROM
        lesson_category
      WHERE
        id IN(?)
      UNION ALL
      SELECT
        cd.id,
        cd.parent_id,
        cd.name,
        cd.sort
      FROM
        lesson_category cd
        INNER JOIN cte ON cd.parent_id = cte.id
    )
    SELECT * FROM cte;
    `
	err := l.db.WithContext(ctx).Raw(rawSQL, parentIDs).Scan(&descendants).Error
	if err != nil {
		return nil, err
	}
	categoryIds := make([]int64, 0)
	lo.ForEach(descendants, func(item *model.LessonCategory, index int) {
		categoryIds = append(categoryIds, item.ID)
	})
	return categoryIds, err
}

func (l *Lesson) UpdateCategorySort(ctx context.Context, sortList []*do.UpdateSort) error {
	qs := query.Use(l.db).LessonCategory
	return l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, v := range sortList {
			_, err := qs.WithContext(ctx).Where(qs.ID.Eq(v.ID)).Update(qs.Sort, v.Sort)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (l *Lesson) MoveLesson(ctx context.Context, req *do.MoveLesson) error {
	qs := query.Use(l.db).Lesson
	return l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		_, err := qs.WithContext(ctx).Where(qs.ID.In(req.LessonIds...)).
			UpdateSimple(
				qs.CategoryID.Value(req.CategoryID),
				qs.UpdateBy.Value(req.UserID),
				qs.UpdateAt.Value(time.Now()))
		if err != nil {
			return err
		}
		return nil
	})
}

func (l *Lesson) GetLessonById(ctx context.Context, id int64) (*model.Lesson, error) {
	qs := query.Use(l.db).Lesson
	return qs.WithContext(ctx).Where(qs.ID.Eq(id)).First()
}
func (l *Lesson) GetLessonByIds(ctx context.Context, ids []int64) (map[int64]*model.Lesson, error) {
	qs := query.Use(l.db).Lesson
	list, err := qs.WithContext(ctx).Where(qs.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	return lo.SliceToMap(list, func(item *model.Lesson) (int64, *model.Lesson) {
		return item.ID, item
	}), nil
}
func (l *Lesson) UpdateLesson(ctx context.Context, req *do.UpdateLesson) error {
	qs := query.Use(l.db).Lesson
	_, err := qs.WithContext(ctx).Where(qs.ID.Eq(req.ID)).UpdateSimple(
		qs.Name.Value(req.Name),
		qs.Detail.Value(req.Detail),
		qs.VideoKey.Value(req.VideoKey),
		qs.VideoFileName.Value(req.VideoFileName),
		qs.Duration.Value(req.Duration),
		qs.Attachments.Value(gconv.String(req.Attachments)),
		qs.Chapters.Value(gconv.String(req.Chapters)),
		qs.UpdateAt.Value(time.Now()),
		qs.UpdateBy.Value(req.UserID),
	)
	return err
}
func (l *Lesson) UpdateLessonStatus(ctx context.Context, req *do.UpdateLessonStatus) error {
	qs := query.Use(l.db).Lesson
	_, err := qs.WithContext(ctx).Where(qs.ID.Eq(req.ID)).Update(qs.Status, req.Status)
	return err
}

func (l *Lesson) CreateLesson(ctx context.Context, req *do.CreateLesson) (int64, error) {
	timeNow := time.Now()
	addObj := &model.Lesson{
		Name:          req.Name,
		Detail:        req.Detail,
		CategoryID:    req.CategoryID,
		VideoKey:      req.VideoKey,
		Attachments:   gconv.String(req.Attachments),
		Duration:      req.Duration,
		VideoFileName: req.VideoFileName,
		Chapters:      gconv.String(req.Chapters),
		Status:        consts.IsEnable,
		CreateBy:      req.UserID,
		UpdateBy:      req.UserID,
		CreateAt:      timeNow,
		UpdateAt:      timeNow,
	}
	err := l.db.WithContext(ctx).Create(addObj).Error
	return addObj.ID, err
}

func (l *Lesson) LessonList(ctx context.Context, req *do.ListLesson) ([]*model.Lesson, int64, error) {
	qs := query.Use(l.db).Lesson
	tx := qs.WithContext(ctx)
	if req.ID > 0 {
		tx = tx.Where(qs.ID.Eq(req.ID))
	}
	if len(req.ExcludeIds) > 0 {
		tx = tx.Where(qs.ID.NotIn(req.ExcludeIds...))
	}
	if req.NameKw != "" {
		tx = tx.Where(qs.Name.Like(tools.GetAllLike(req.NameKw)))
	}
	if req.CategoryIDs != nil && len(req.CategoryIDs) > 0 {
		tx = tx.Where(qs.CategoryID.In(req.CategoryIDs...))
	}
	if req.Status != 0 {
		tx = tx.Where(qs.Status.Eq(req.Status))
	}
	if req.StartCreateTime > 0 && req.EndCreateTime > 0 {
		tx = tx.Where(qs.CreateAt.Between(time.UnixMilli(req.StartCreateTime), time.UnixMilli(req.EndCreateTime)))
	}
	if req.StartUpdateTime > 0 && req.EndUpdateTime > 0 {
		tx = tx.Where(qs.CreateAt.Between(time.UnixMilli(req.StartUpdateTime), time.UnixMilli(req.EndUpdateTime)))
	}
	return tx.Order(qs.Status.Desc(), qs.ID.Desc()).FindByPage(req.GetOffset(), req.Limit)
}
