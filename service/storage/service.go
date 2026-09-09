package storage

import (
	"mall/adaptor"
	"mall/adaptor/repo/upload"
	"mall/adaptor/rpc"
)

type Service struct {
	cos  rpc.IStorage
	repo upload.IUploadFile
}

func NewService(adaptor adaptor.IAdaptor) *Service {
	return &Service{
		cos:  rpc.NewStorage(adaptor),
		repo: upload.NewUploadFile(adaptor),
	}
}
