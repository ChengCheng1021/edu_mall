package do

type LessonLearnBase struct {
	UserID       int64
	CourseID     int64
	LessonID     int64
	PlayPosition int64
	LearnStatus  int32
}

type LessonLearnSessionInfo struct {
	CourseID         int64
	LessonID         int64
	FirstCategoryID  int64
	SecondCategoryID int64
	LessonDuration   int64
}
