package models

type SignUpDTO struct {
	Type      string `json:"type" binding:"required"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	Tel       string `json:"tel"`
	WxUnionId string `json:"wxUnionId"`
	NickName  string `json:"nickName"`
	Icon      string `json:"icon"`
	Location  string `json:"location"`
}

type LoginDTO struct {
	LoginName string `json:"loginName" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type TokenRefreshDTO struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type PasswordChangeDTO struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type ProfileUpdateDTO struct {
	NickName string `json:"nickName"`
	Icon     string `json:"icon"`
	Gender   string `json:"gender"`
	Birth    string `json:"birth"`
	Location string `json:"location"`
}

type ThirdLoginDTO struct {
	Code        string `json:"code" binding:"required"`
	RedirectUri string `json:"redirectUri"`
	State       string `json:"state"`
}

type WxMiniLoginDTO struct {
	Code     string `json:"code" binding:"required"`
	AppId    string `json:"appId" binding:"required"`
	NickName string `json:"nickName"`
	Icon     string `json:"icon"`
	Gender   string `json:"gender"`
}
