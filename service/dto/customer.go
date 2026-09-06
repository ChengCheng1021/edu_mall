package dto

import "mall/common"

type CustomerUserDto struct {
	UserDto
	Mobile string `json:"mobile"`
}

type CustomerUserInfoDto struct {
	*UserInfoDto
	UserCourses []*PurchasedCourseDto `json:"user_courses"`
}

type ListCustomerUserReq struct {
	common.Pager
	ID         int64  `form:"id"`
	NickNameKw string `form:"nick_name_kw"`
	Status     int32  `form:"status"`
	Mobile     string `form:"mobile"`
}

type ListCustomerUserResp struct {
	List  []*CustomerUserDto `json:"list"`
	Total int64              `json:"total"`
	common.Pager
}

type GetCustomerUserInfoReq struct {
	ID int64 `form:"id"`
}
