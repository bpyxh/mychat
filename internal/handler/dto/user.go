package dto

type Response struct {
	Code int    `json:"code,omitempty" example:"200"`
	Msg  string `json:"msg" example:"操作成功"`
	Data any    `json:"data,omitempty"`
}

type ErrorResp struct {
	Code int    `json:"code,omitempty" example:"101000"`
	Msg  string `json:"msg" example:"操作失败"`
	Data any    `json:"data,omitempty"`
}

type CreateUserReq struct {
	UserName  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Password2 string `json:"password2" binding:"required"`
	Name      string `json:"name"`
}

type CreateUserResp struct {
	Id string `json:"id"`
}

type UserResp struct {
	ID       uint32 `json:"id"`
	UserName string `json:"username"`
	Name     string `json:"name"`
}

type LoginReq struct {
	UserName string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type LoginRespData struct {
	Expire string   `json:"expire"`
	Token  string   `json:"token"`
	User   UserResp `json:"user"`
}

type LoginResp struct {
	Response
	Data LoginRespData `json:"data"`
}
