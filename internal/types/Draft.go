package types

import "blog/internal/model"

type DraftSaveRequest struct {
	Title      string `json:"title" binding:"required,max=255"`
	ContentURL string `json:"content_url"`
	Content    string `json:"content"` // 兼容旧字段，等价于 content_url
}

type DraftIDResponse struct {
	ID int64 `json:"id"`
}

type DraftListResponse struct {
	List []*model.Draft `json:"list"`
}
