package models

type LoginVO struct {
	Sn           string  `json:"sn"`
	Token        string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	TokenType    string  `json:"token_type"`
	ExpiresIn    int64   `json:"expires_in"`
	User         *UserVO `json:"user"`
}

type RegisterVO struct {
	Id        uint64 `json:"id"`
	Sn        string `json:"sn"`
	Username  string `json:"username"`
	Token     string `json:"access_token"`
	TokenType string `json:"token_type"`
	ExpiresIn int64  `json:"expires_in"`
}

type UserVO struct {
	Id                   uint64 `json:"id"`
	Sn                   string `json:"sn"`
	Name                 string `json:"name"`
	NickName             string `json:"nickName"`
	Icon                 string `json:"icon"`
	Gender               string `json:"gender"`
	Birth                string `json:"birth"`
	CreateFrom           string `json:"createFrom"`
	Location             string `json:"location"`
	City                 string `json:"city"`
	WxUnionId            string `json:"wxUnionId"`
	Email                string `json:"email"`
	Tel                  string `json:"tel"`
	CreateAt             int64  `json:"createAt"`
	AccountNonExpired    int    `json:"accountNonExpired"`
	AccountNonLocked     int    `json:"accountNonLocked"`
	CredentialsNonExpired int   `json:"credentialsNonExpired"`
	Enabled              int    `json:"enabled"`
}

type TokenVerifyVO struct {
	Sn       string   `json:"sn"`
	Name     string   `json:"name"`
	NickName string   `json:"nickName"`
	Icon     string   `json:"icon"`
	Email    string   `json:"email"`
	Tel      string   `json:"tel"`
	Gender   string   `json:"gender"`
	Location string   `json:"location"`
	CreateFrom string `json:"createFrom"`
}

type TokenRefreshVO struct {
	Sn           string `json:"sn"`
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type MiniLoginVO struct {
	Sn           string  `json:"sn"`
	XToken       string  `json:"x_token"`
	XRefreshToken string `json:"x_refresh_token"`
	AppToken     string  `json:"access_token"`
	TokenType    string  `json:"token_type"`
	ExpiresIn    int64   `json:"expires_in"`
	User         *UserVO `json:"user"`
}
