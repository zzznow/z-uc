package auth

import (
	LOGGER "log/slog"
	"net/http"
	"net/url"

	"github.com/zzznow/z-uc/models"
	"github.com/gin-gonic/gin"
)

type WxAccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionId      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

type WxUserInfo struct {
	OpenId     string `json:"openid"`
	Nickname   string `json:"nickname"`
	Sex        int    `json:"sex"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	HeadImgUrl string `json:"headimgurl"`
	UnionId    string `json:"unionid"`
}

func WxToken(c *gin.Context) {
	var req models.ThirdLoginDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	savedState, err := getState(c, req.State)
	if err != nil || savedState == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	wxResp, err := exchangeWxCode(req.Code)
	if err != nil {
		LOGGER.Error("wechat token exchange failed", "err", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "wechat auth failed"})
		return
	}

	if wxResp.ErrCode != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": wxResp.ErrMsg})
		return
	}

	wxUserInfo, err := getWxUserInfo(wxResp.AccessToken, wxResp.OpenId)
	if err != nil {
		LOGGER.Error("wechat userinfo failed", "err", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "wechat auth failed"})
		return
	}

	unionId := wxUserInfo.UnionId
	if unionId == "" {
		unionId = wxResp.UnionId
	}
	if unionId == "" {
		unionId = wxResp.OpenId
	}

	user, err := signUpOrLoginByThird("wx", unionId, unionId, unionId+"-wx",
		wxUserInfo.Nickname, wxUserInfo.HeadImgUrl, "", wxUserInfo.Country+wxUserInfo.Province+wxUserInfo.City)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	respondWithTokens(c, user)
}

func exchangeWxCode(code string) (*WxAccessTokenResponse, error) {
	params := url.Values{}
	params.Set("appid", "wx_appid")
	params.Set("secret", "wx_secret")
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")

	resp, err := restyClient.R().
		SetResult(&WxAccessTokenResponse{}).
		Get("https://api.weixin.qq.com/sns/oauth2/access_token?" + params.Encode())

	if err != nil {
		return nil, err
	}

	return resp.Result().(*WxAccessTokenResponse), nil
}

func getWxUserInfo(accessToken, openId string) (*WxUserInfo, error) {
	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("openid", openId)
	params.Set("lang", "zh_CN")

	resp, err := restyClient.R().
		SetResult(&WxUserInfo{}).
		Get("https://api.weixin.qq.com/sns/userinfo?" + params.Encode())

	if err != nil {
		return nil, err
	}

	return resp.Result().(*WxUserInfo), nil
}
