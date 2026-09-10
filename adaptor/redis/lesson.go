package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mall/adaptor"
	"mall/config"
	"time"

	"github.com/go-redis/redis"
)

type ILessonLearn interface {
	GetLessonLearnSession(ctx context.Context, userID, courseID, lessonID int64) (*LessonLearnSession, error)
}

type LessonLearnSession struct {
	UserID             int64 `json:"user_id"`
	CourseID           int64 `json:"course_id"`
	LessonID           int64 `json:"lesson_id"`
	FirstCategoryID    int64 `json:"first_category_id"`
	SecondCategoryID   int64 `json:"second_category_id"`
	LessonDuration     int64 `json:"lesson_duration"`
	EntryTime          int64 `json:"entry_time"`
	LastEventTime      int64 `json:"last_event_time"`
	LastPlayTime       int64 `json:"last_play_time"`
	AccumulateDuration int64 `json:"accumulate_duration"`
	PlayPosition       int64 `json:"play_position"`
	LastType           int32 `json:"last_type"`
}

type LessonLearn struct {
	redis *redis.Client
}

const (
	lessonLearnTimeoutBatchSize = 500
	lessonLearnLockExpire       = time.Second * 15
)

func NewLessonLearn(adaptor adaptor.IAdaptor) *LessonLearn {
	return &LessonLearn{redis: adaptor.GetRedis()}
}

func fmtLessonLearnSessionKey(userID, courseID, lessonID int64) string {
	return fmt.Sprintf("%s:lesson:learn:session:%d:%d:%d", config.ServerFullName, userID, courseID, lessonID)
}

func (l *LessonLearn) GetLessonLearnSession(ctx context.Context, userID, courseID, lessonID int64) (*LessonLearnSession, error) {
	value, err := l.redis.Get(fmtLessonLearnSessionKey(userID, courseID, lessonID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	session := &LessonLearnSession{}
	if err = json.Unmarshal([]byte(value), session); err != nil {
		return nil, err
	}
	return session, nil
}
