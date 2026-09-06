package role

import (
	"mall/adaptor"
	"mall/adaptor/repo/admin"
)

type Service struct {
	adminRole admin.IAdminRole
	adminPerm admin.IPerm
}

func NewService(adaptor adaptor.IAdaptor) *Service {
	return &Service{
		adminRole: admin.NewAdminRole(adaptor),
		adminPerm: admin.NewAdminPerm(adaptor),
	}
}
