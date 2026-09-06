package dto

import "mall/common"

type AddCategoryReq struct {
	Name     string `json:"name"`
	ParentID int64  `json:"parent_id"`
	Level    int32  `json:"level"`
	Sort     int32  `json:"sort"`
}

type UpdateCategoryReq struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type DeleteCategoryReq struct {
	IDs []int64 `json:"ids"`
}

type UpdateSort struct {
	ID       int64 `json:"id"`
	Sort     int32 `json:"sort"`
	Level    int32 `json:"level"`
	ParentID int64 `json:"parent_id"`
}

type CategoryDto struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Sort     int32  `json:"sort"`
	ParentID int64  `json:"parent_id"`
	Level    int32  `json:"level"`
}

type ListCategoryReq struct {
	common.Pager
}

type LessonChapter struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	BeginPosition int64  `json:"begin_position"`
	EndPosition   int64  `json:"end_position"`
}

type Attachment struct {
	OriginName string `json:"origin_name"`
	FileKey    string `json:"file_key"`
	FileUrl    string `json:"file_url"`
}

type CreateLessonReq struct {
	Name          string          `json:"name"`
	Detail        string          `json:"detail"`
	CategoryID    int64           `json:"category_id"`
	VideoKey      string          `json:"video_key"`
	Attachments   []Attachment    `json:"attachments"`
	Duration      int32           `json:"duration"`
	Chapters      []LessonChapter `json:"chapters"`
	VideoFileName string          `json:"video_file_name"`
}

type ListLessonReq struct {
	common.Pager
	CourseID        int64  `json:"course_id"`
	ID              int64  `json:"id"`
	OnView          bool   `json:"on_view"`
	NameKw          string `json:"name_kw"`
	CategoryID      int64  `json:"category_id"`
	Status          int32  `json:"status"`
	StartCreateTime int64  `json:"start_create_time"`
	EndCreateTime   int64  `json:"end_create_time"`
	StartUpdateTime int64  `json:"begin_update_time"`
	EndUpdateTime   int64  `json:"end_update_time"`
}

type LessonDto struct {
	common.CreateUpdateName
	ID            int64           `json:"id"`
	Name          string          `json:"name"`
	Detail        string          `json:"detail"`
	CategoryID    int64           `json:"category_id"`
	CategoryName  string          `json:"category_name"`
	VideoKey      string          `json:"video_key"`
	VideoUrl      string          `json:"video_url"`
	VideoFileName string          `json:"video_file_name"`
	Attachments   []Attachment    `json:"attachments"`
	Duration      int32           `json:"duration"`
	Chapters      []LessonChapter `json:"chapters"`
	Status        int32           `json:"status"`
	CreateBy      int64           `json:"create_by"`
	UpdateBy      int64           `json:"update_by"`
	CreateAt      int64           `json:"create_at"`
	UpdateAt      int64           `json:"update_at"`
}

type ListLessonResp struct {
	List  []*LessonDto `json:"list"`
	Total int64        `json:"total"`
	common.Pager
}

type LessonInfoReq struct {
	ID       int64 `form:"id"`
	LessonID int64 `form:"lesson_id"`
	UserId   int64 `form:"user_id"`
}

type UpdateLessonReq struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Detail      string          `json:"detail"`
	VideoKey    string          `json:"video_key"`
	Attachments []Attachment    `json:"attachments"`
	Duration    int32           `json:"duration"`
	Chapters    []LessonChapter `json:"chapters"`
}

type UpdateLessonStatusReq struct {
	ID     int64 `json:"id"`
	Status int32 `json:"status"`
}

type MoveLessonReq struct {
	LessonIds  []int64 `json:"lesson_ids"`
	CategoryID int64   `json:"category_id"`
}
