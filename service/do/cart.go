package do

import "mall/common"

type AddGoods struct {
	GoodsID int64
	UserID  int64
}

type RemoveGoods struct {
	ID     int64
	UserID int64
}

type ListGoods struct {
	common.Pager
	GoodsNameKW string
	UserID      int64
}
