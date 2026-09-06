package dto

import "mall/common"

type CourseFileDto struct {
	Title    string `json:"title"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
	FileKey  string `json:"file_key,omitempty"`
	Url      string `json:"url,omitempty"`
}

type CreateCourseReq struct {
	Name           string           `json:"name"`
	CoursePrice    int64            `json:"course_price"`
	ServiceTime    int32            `json:"service_time"`
	LearnTime      int32            `json:"learn_time"`
	Sort           int32            `json:"sort"`
	Features       []string         `json:"features"`
	UpdateStatus   int32            `json:"update_status"`
	CoverKey       string           `json:"cover_key"`
	DetailCoverKey string           `json:"detail_cover_key"`
	Detail         string           `json:"detail"`
	Files          []*CourseFileDto `json:"files"`
}

type CourseDto struct {
	common.CreateUpdateName
	ID                int64            `json:"id"`
	Name              string           `json:"name"`
	CoursePrice       int64            `json:"course_price"`
	ServiceTime       int32            `json:"service_time"`
	LearnTime         int32            `json:"learn_time"`
	Sort              int32            `json:"sort"`
	Status            int32            `json:"status"`
	Features          []string         `json:"features"`
	UpdateStatus      int32            `json:"update_status"`
	HasPurchased      bool             `json:"has_purchased"`
	ServiceExpireTime int64            `json:"service_expire_time"`
	LearnExpireTime   int64            `json:"learn_expire_time"`
	CoverKey          string           `json:"cover_key"`
	CoverUrl          string           `json:"cover_url"`
	DetailCoverKey    string           `json:"detail_cover_key"`
	DetailCoverUrl    string           `json:"detail_cover_url"`
	Detail            string           `json:"detail"`
	Files             []*CourseFileDto `json:"files"`
	CreateBy          int64            `json:"create_by"`
	UpdateBy          int64            `json:"update_by"`
	CreateAt          int64            `json:"create_at"`
	UpdateAt          int64            `json:"update_at"`
}

type PurchasedCourseDto struct {
	ID                int64            `json:"id"`
	Name              string           `json:"name"`
	ServiceExpireTime int64            `json:"service_expire_time"`
	LearnExpireTime   int64            `json:"learn_expire_time"`
	Features          []string         `json:"features"`
	UpdateStatus      int32            `json:"update_status"`
	HasPurchased      bool             `json:"has_purchased"`
	CoverKey          string           `json:"cover_key"`
	CoverUrl          string           `json:"cover_url"`
	DetailCoverKey    string           `json:"detail_cover_key"`
	DetailCoverUrl    string           `json:"detail_cover_url"`
	Detail            string           `json:"detail"`
	Files             []*CourseFileDto `json:"files"`
}

type CourseInfoReq struct {
	ID      int64 `form:"id"`
	UserID  int64 `json:"-" form:"-"`
	IsAdmin bool  `json:"-" form:"-"`
}

type UpdateCourseReq struct {
	ID             int64            `json:"id"`
	Name           string           `json:"name"`
	CoursePrice    int64            `json:"course_price"`
	ServiceTime    int32            `json:"service_time"`
	LearnTime      int32            `json:"learn_time"`
	Sort           int32            `json:"sort"`
	Features       []string         `json:"features"`
	UpdateStatus   int32            `json:"update_status"`
	CoverKey       string           `json:"cover_key"`
	DetailCoverKey string           `json:"detail_cover_key"`
	Detail         string           `json:"detail"`
	Files          []*CourseFileDto `json:"files"`
}

type UpdateCourseStatusReq struct {
	ID     int64 `json:"id"`
	Status int32 `json:"status"`
}

type CourseListReq struct {
	common.Pager
	UserID          int64  `form:"user_id"`
	ID              int64  `form:"id"`
	NameKw          string `form:"name_kw"`
	CreateStartTime int64  `form:"create_start_time"`
	CreateEndTime   int64  `form:"create_end_time"`
	UpdateStartTime int64  `form:"update_start_time"`
	UpdateEndTime   int64  `form:"update_end_time"`
	UpdateStatus    int32  `form:"update_status"`
	Status          int32  `form:"status"`
	IsRecommend     bool   `form:"is_recommend"`
	IsAdmin         bool   `json:"-" form:"-"`
}

type CourseListResp struct {
	List  []*CourseDto `json:"list"`
	Total int64        `json:"total"`
	common.Pager
}

type AddCatalogReq struct {
	CourseID int64  `json:"course_id"`
	Name     string `json:"name"`
	ParentID int64  `json:"parent_id"`
	Sort     int32  `json:"sort"`
	Level    int32  `json:"level"`
}

type UpdateCatalogReq struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type DeleteCatalogReq struct {
	ID int64 `json:"id"`
}

type CatalogInfoReq struct {
	CourseID    int64 `form:"course_id"`
	FilterVideo bool
}

type CatalogLessonDto struct {
	ID            int64  `json:"id"`
	LessonID      int64  `json:"lesson_id"`
	Index         int    `json:"index"`
	Name          string `json:"name"`
	LessonName    string `json:"lesson_name"`
	Detail        string `json:"detail"`
	VideoUrl      string `json:"video_url"`
	VideoFileName string `json:"video_file_name"`
	Duration      int32  `json:"duration"`
	Status        int32  `json:"status"`
	ShowTime      int64  `json:"show_time"`
	EnableTrial   int32  `json:"enable_trial"`
}

type CatalogDto struct {
	CourseID    int64               `json:"course_id"`
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	Sort        int32               `json:"sort"`
	ParentID    int64               `json:"parent_id"`
	Level       int32               `json:"level"`
	Lessons     []*CatalogLessonDto `json:"lessons"`
	LessonCount int64               `json:"lesson_count"`
}

type CatalogInfoResp struct {
	TotalDuration int64         `json:"total_duration"`
	LessonCount   int64         `json:"lesson_count"`
	Catalogs      []*CatalogDto `json:"catalogs"`
}

type AddCatalogLessonReq struct {
	CourseID  int64   `json:"course_id"`
	CatalogID int64   `json:"catalog_id"`
	LessonIDs []int64 `json:"lesson_ids"`
}

type UpdateCatalogLessonReq struct {
	ID          int64  `json:"id"`
	EnableTrial int32  `json:"enable_trial"`
	Name        string `json:"name	"`
	ShowTime    int64  `json:"show_time"`
}

type RemoveCatalogLessonReq struct {
	IDs []int64 `json:"ids"`
}

type UpdateCatalogSortDto struct {
	Id       int64            `json:"id"`
	Sort     int32            `json:"sort"`
	ParentId int64            `json:"parent_id"`
	Level    int32            `json:"level"`
	Lessons  []*common.IDSort `json:"lessons"`
}

type CourseDetailDto struct {
	*CourseDto
	*CatalogInfoResp
}
