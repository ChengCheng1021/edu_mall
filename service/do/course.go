package do

import "mall/common"

type CourseFile struct {
	Title    string `json:"title"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
	FileKey  string `json:"file_key"`
}

type CreateCourse struct {
	UserID         int64
	Name           string
	CoursePrice    int64
	ServiceTime    int32
	LearnTime      int32
	Sort           int32
	Features       []string
	UpdateStatus   int32
	CoverKey       string
	DetailCoverKey string
	Detail         string
	Files          []*CourseFile
}

type UpdateCourse struct {
	UserID         int64
	ID             int64
	Name           string
	CoursePrice    int64
	ServiceTime    int32
	LearnTime      int32
	Sort           int32
	Features       []string
	UpdateStatus   int32
	CoverKey       string
	DetailCoverKey string
	Detail         string
	Files          []*CourseFile
}

type UpdateCourseStatus struct {
	UserID int64
	ID     int64
	Status int32
}

type CourseList struct {
	common.Pager
	ID                    int64
	UserID                int64
	NameKw                string
	StartCreateTime       int64
	EndCreateTime         int64
	StartUpdateTime       int64
	EndUpdateTime         int64
	UpdateStatus          int32
	Status                int32
	ExcludePurchasedGoods bool
}

type AddCatalog struct {
	CourseID int64
	Name     string
	ParentID int64
	Sort     int32
	Level    int32
	UserID   int64
}

type UpdateCatalog struct {
	ID     int64
	UserID int64
	Name   string
}

type DeleteCatalog struct {
	ID int64
}

type AddCatalogLesson struct {
	UserID    int64
	CourseID  int64
	CatalogID int64
	LessonMap map[int64]string
	LessonIDs []int64
}

type UpdateCatalogLesson struct {
	ID          int64
	UserID      int64
	EnableTrial int32
	Name        string
	ShowTime    int64
}
type RemoveCatalogLesson struct {
	IDs []int64
}

type CatalogSort struct {
	Id       int64
	Sort     int32
	ParentId int64
	Level    int32
	Lessons  []*common.IDSort
}

type UpdateCatalogSort struct {
	SortList []*CatalogSort
	UserID   int64
}
