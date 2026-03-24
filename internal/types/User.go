package types

type UserInfoResponse struct {
	ID       int    `json:"id" example:"1"`
	Username string `json:"username" example:"alice"`
	Image    string `json:"image" example:"/uploads/1/avatar.png"`
	Age      int32  `json:"age" example:"25"`
	Gender   string `json:"gender" example:"female"`
}
