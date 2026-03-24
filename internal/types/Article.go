package types

type CreateArticleRequest struct {
	Title      string `json:"title"`
	ContentURL string `json:"content_url"` // markdown 文件 URL
	Content    string `json:"content"`     // 兼容旧字段，等价于 content_url
	DraftID    int    `json:"draft_id"`    // 发布时可选择草稿
	Status     string `json:"status"`      // draft/published/offline
	CategoryID int    `json:"category_id"`
	Tags       string `json:"tags"`      // 标签，逗号分隔
	CoverURL   string `json:"cover_url"` // 封面图URL
}

type UpdateArticleRequest struct {
	Title      string `json:"title"`
	ContentURL string `json:"content_url"` // markdown 文件 URL
	Content    string `json:"content"`     // 兼容旧字段，等价于 content_url
	CategoryID int    `json:"category_id"`
	Tags       string `json:"tags"`
	CoverURL   string `json:"cover_url"`
}

type UpdateArticleStatusRequest struct {
	Status string `json:"status"` // draft/published/offline
}

type CreateArticleResponse struct {
	ID int64 `json:"id"`
}
