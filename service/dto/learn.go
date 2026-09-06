package dto

import "mall/common"

type LessonLearnInfoReq struct {
	CourseID int64 `form:"course_id"`
	LessonID int64 `form:"lesson_id"`
}

type LessonLearnInfoResp struct {
	CourseID       int64 `json:"course_id"`
	LessonID       int64 `json:"lesson_id"`
	PlayPosition   int64 `json:"play_position"`
	LearnStatus    int32 `json:"learn_status"`
	LastType       int32 `json:"last_type"`
	EntryTime      int64 `json:"entry_time"`
	LastReportTime int64 `json:"last_report_time"`
	InLearning     bool  `json:"in_learning"`
}

type LessonLearnReportReq struct {
	CourseID     int64 `json:"course_id"`
	LessonID     int64 `json:"lesson_id"`
	Type         int32 `json:"type"`
	PlayPosition int64 `json:"play_position"`
}

type ContinueLearnReq struct {
	common.Pager
}

type ContinueLearnCourseDto struct {
	CourseID       int64  `json:"course_id"`
	CourseName     string `json:"course_name"`
	CourseCoverKey string `json:"course_cover_key"`
	CourseCoverUrl string `json:"course_cover_url"`
	LessonID       int64  `json:"lesson_id"`
	LessonName     string `json:"lesson_name"`
	LessonIndex    int64  `json:"lesson_index"`
	LessonCount    int64  `json:"lesson_count"`
	PlayPosition   int64  `json:"play_position"`
	LearnStatus    int32  `json:"learn_status"`
	LastLearnTime  int64  `json:"last_learn_time"`
}

type ContinueLearnResp struct {
	List  []*ContinueLearnCourseDto `json:"list"`
	Total int64                     `json:"total"`
	common.Pager
}
