package dto

import "mall/common"

type AddGoodsReq struct {
	GoodsID int64 `json:"goods_id"`
}

type RemoveGoodsReq struct {
	ID int64 `json:"id"`
}

type ListGoodsReq struct {
	common.Pager
	GoodsNameKw string `json:"goods_name_kw"`
}

type CartGoodsDto struct {
	ID       int64 `json:"cart_id"`
	GoodsID  int64 `json:"goods_id"`
	Quantity int32 `json:"quantity"`
	*CourseDto
}

type ListGoodsResp struct {
	common.Pager
	Total int64           `json:"total"`
	List  []*CartGoodsDto `json:"list"`
}
