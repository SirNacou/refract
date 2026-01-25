package dto

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

type ListURLsParams struct {
	Page   int    `json:"page" validate:"min=1"`
	Limit  int    `json:"limit" validate:"min=1,max=100"`
	Status string `json:"status" validate:"omitempty,oneof=active expired disabled deleted"`
	Search string `json:"search" validate:"omitempty,max=200"`
	Sort   string `json:"sort" validate:"omitempty,oneof=created_at updated_at total_clicks"`
	Order  string `json:"order" validate:"omitempty,oneof=asc desc"`
}
