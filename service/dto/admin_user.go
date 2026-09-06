package dto

import "mall/common"

type AdminUserDto struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	Name       string `json:"name"`
	NickName   string `json:"nick_name"`
	Sex        int32  `json:"sex"`
	Status     int32  `json:"status"`
	Mobile     string `json:"mobile"`
	LarkOpenID string `json:"lark_open_id"`
	UpdateAt   int64  `json:"update_at"`
	CreateAt   int64  `json:"create_at"`
}

type CreateUserReq struct {
	Name     string  `json:"name"`
	NickName string  `json:"nick_name"`
	Mobile   string  `json:"mobile"`
	Sex      int32   `json:"sex"`
	RoleIDs  []int64 `json:"role_ids"`
}

type UpdateUserReq struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	NickName string  `json:"nick_name"`
	Sex      int32   `json:"sex"`
	Status   int32   `json:"status"`
	RoleIDs  []int64 `json:"role_ids"`
}

type DeleteUserReq struct {
	ID int64 `json:"id"`
}

type ListAdminUserReq struct {
	common.Pager
	Name   string `form:"name_kw,omitempty"`
	Mobile string `form:"mobile,omitempty"`
	RoleID int64  `form:"role_id,omitempty"`
	Status int32  `form:"status,omitempty"`
}

type AdminUserWithRoleDto struct {
	AdminUserDto
	Roles []*common.IDName `json:"roles"`
}

type ListAdminUserResp struct {
	common.Pager
	Total int64                   `json:"total"`
	List  []*AdminUserWithRoleDto `json:"list"`
}

type LarkQrCodeBindReq struct {
	AppCode     int32  `json:"app_code"`
	Code        string `json:"code"`
	RedirectUri string `json:"redirect_uri"`
}

type ChangeMobileReq struct {
	Mobile     string `json:"mobile"`
	VerifyCode string `json:"verify_code"`
}
