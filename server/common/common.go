package common

type RequestSearch struct {
	Keyword string `json:"keyword" form:"keyword"`
	OrderBy string `json:"order_by" form:"order_by"`
}
