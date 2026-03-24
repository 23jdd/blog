package types

type CreateCommentRequest struct {
	Content  string `json:"content"`
	ParentID int    `json:"parent_id"`
}

type ReviewStatusRequest struct {
	Status string `json:"status"` // pending/approved/rejected
}

type LikeActionResponse struct {
	Liked bool `json:"liked"`
}
