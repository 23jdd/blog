package types

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type LoginResponse struct {
	Token string `json:"token"`
}
type ErrorResponse struct {
	Message string `json:"message"` // 错误信息
	Token   string `json:"token"`   //  token
}
