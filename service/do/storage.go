package do

type AddUploadFile struct {
	Scene    string
	FileKey  string
	FileName string
	FileSize int64
	FileType string
	ClientIP string
	UserID   int64
	UserType int32
}

type GetTempSecret struct {
	Scene    string
	FileName string
	FileSize int64
	FileType string
	ClientIP string
}

type TempSecret struct {
	SecretId      string
	SecretKey     string
	SecurityToken string
	ExpiredTime   int64
	StartTime     int64
	Bucket        string
	Region        string
	Key           string
	FileUrl       string
}

type GetPreviewUrl struct {
	Keys        []string
	ExpireHours int
}

type DeleteFile struct {
	Keys []string
}
