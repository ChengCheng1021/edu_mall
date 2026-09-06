package do

import "mall/common"

type AddCategory struct {
	Name     string
	Level    int32
	ParentID int64
	Sort     int32
}

type UpdateCategory struct {
	ID   int64
	Name string
}

type DeleteCategory struct {
	IDs []int64
}

type UpdateSort struct {
	ID   int64
	Sort int32
}

type UpdateCategorySort []UpdateSort

type ListCategory struct {
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
}

type CreateLesson struct {
	UserID        int64
	Name          string
	Detail        string
	CategoryID    int64
	VideoKey      string
	Attachments   []Attachment
	Duration      int32
	VideoFileName string
	Chapters      []LessonChapter
}

type ListLesson struct {
	common.Pager
	ExcludeIds      []int64
	ID              int64
	NameKw          string
	CategoryIDs     []int64
	Status          int32
	StartCreateTime int64
	EndCreateTime   int64
	StartUpdateTime int64
	EndUpdateTime   int64
}

type UpdateLesson struct {
	UserID        int64
	ID            int64
	Name          string
	Detail        string
	VideoKey      string
	VideoFileName string
	Attachments   []Attachment
	Duration      int32
	Chapters      []LessonChapter
}

type UpdateLessonStatus struct {
	UserID int64
	ID     int64
	Status int32
}

type MoveLesson struct {
	UserID     int64
	LessonIds  []int64
	CategoryID int64
}
