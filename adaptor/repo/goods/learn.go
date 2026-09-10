package goods

import (
	"context"
	"errors"
	"mall/adaptor"
	"mall/adaptor/repo/model"
	"mall/adaptor/repo/query"

	"gorm.io/gorm"
)

type ILearn interface {
	GetLessonLearnProgress(ctx context.Context, userID, courseID, lessonID int64) (*model.LessonLearnProgress, error)
}

type Learn struct {
	db *gorm.DB
}

func NewLearn(adaptor adaptor.IAdaptor) *Learn {

	return &Learn{db: adaptor.GetDB()}
}
func (l *Learn) GetLessonLearnProgress(ctx context.Context, userID, courseID, lessonID int64) (*model.LessonLearnProgress, error) {
	qs := query.Use(l.db).LessonLearnProgress
	obj, err := qs.WithContext(ctx).Where(
		qs.UserID.Eq(userID),
		qs.CourseID.Eq(courseID),
		qs.LessonID.Eq(lessonID),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return obj, nil
}
