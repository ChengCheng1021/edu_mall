package do

type AddPerm struct {
	AdminUserID int64
	Code        string
	Name        string
	Type        int32
	Desc        string
	ParentID    int64
	Sort        int32
	PagePath    string
}

type UpdatePerm struct {
	ID int64
	AddPerm
}

type UpdatePermList struct {
	List []UpdatePerm
}

type DeletePerm struct {
	ID int64
}
