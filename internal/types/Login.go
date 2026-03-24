package types

type LoginRequest struct {
	Username string `json:"username" example:"alice"`
	Password string `json:"password" example:"P@ssw0rd123"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type LoginResponse struct {
	Code         int    `json:"code" example:"200"`
	Message      string `json:"message" example:"success"`
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
type ErrorResponse struct {
	Code    int         `json:"code,omitempty" example:"400"` // 错误码
	Message string      `json:"message" example:"参数错误"`       // 错误信息
	Token   string      `json:"token,omitempty" example:""`
	Data    interface{} `json:"data,omitempty" swaggertype:"object"`
}
type SuccessResponse struct {
	Code    int         `json:"code,omitempty" example:"200"` // 业务码
	Message string      `json:"message" example:"操作成功"`       // 成功信息
	Data    interface{} `json:"data,omitempty" swaggertype:"object"`
}
