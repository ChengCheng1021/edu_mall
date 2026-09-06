package common

type Pager struct {
	Page      int  `json:"page" form:"page"`
	Limit     int  `json:"limit" form:"limit"`
	UnLimited bool `json:"unlimited" form:"unlimited"`
}

func (p *Pager) GetOffset() int {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.Limit == 0 {
		p.Limit = 10
	}
	if p.Limit > 1000 && p.UnLimited == false {
		p.Limit = 1000
	}
	return (p.Page - 1) * p.Limit
}

type CreateUpdateName struct {
	CreateName string `json:"create_name,omitempty"`
	UpdateName string `json:"update_name,omitempty"`
}

type IDName struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type IDSort struct {
	ID   int64 `json:"id"`
	Sort int32 `json:"sort"`
}
