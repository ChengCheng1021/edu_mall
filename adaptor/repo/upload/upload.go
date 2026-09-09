package upload

import (
	"context"
	"mall/adaptor"
	"mall/adaptor/repo/model"
	"mall/service/do"

	"gorm.io/gorm"
)

type IUploadFile interface {
	CreateUploadFile(ctx context.Context, fileList []do.AddUploadFile) error
	DeleteUploadFile(ctx context.Context, fileKeys []string) error
}

type UploadFile struct {
	db *gorm.DB
}

func NewUploadFile(adaptor adaptor.IAdaptor) *UploadFile {
	return &UploadFile{
		db: adaptor.GetDB(),
	}
}

func (s *UploadFile) CreateUploadFile(ctx context.Context, fileList []do.AddUploadFile) error {
	var addList []model.ResourceUploadFile
	for _, file := range fileList {
		addList = append(addList, model.ResourceUploadFile{
			Scene:          file.Scene,
			FileKey:        file.FileKey,
			FileName:       file.FileName,
			FileSize:       file.FileSize,
			FileType:       file.FileType,
			UploadClientIP: file.ClientIP,
			UserID:         file.UserID,
			UserType:       file.UserType,
		})
	}
	return s.db.WithContext(ctx).CreateInBatches(&addList, 100).Error
}

func (s *UploadFile) DeleteUploadFile(ctx context.Context, fileKeys []string) error {
	return s.db.WithContext(ctx).Where("file_key in ?", fileKeys).Delete(&model.ResourceUploadFile{}).Error
}
