package auth

import (
	"fmt"
	LOGGER "log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/zzznow/z-uc/models"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"

	"github.com/zzznow/z-uc/auth/internal"
)

type WxMiniSessionResponse struct {
	OpenId     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionId    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type AppTokenResponse struct {
	Token     string `json:"access_token"`
	TokenType string `json:"token_type"`
	ExpiresIn int64  `json:"expires_in"`
}

func WxMiniToken(c *gin.Context) {
	var req models.WxMiniLoginDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wxResp, err := exchangeWxMiniCode(req.Code)
	if err != nil {
		LOGGER.Error("wechat mini program code2session failed", "err", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "wechat auth failed"})
		return
	}

	if wxResp.ErrCode != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": wxResp.ErrMsg})
		return
	}

	unionId := wxResp.UnionId
	if unionId == "" {
		unionId = wxResp.OpenId
	}

	user, err := signUpOrLoginByThird("wxmini", unionId, unionId, unionId+"-wxmp",
		req.NickName, req.Icon, "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	xToken, err := models.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	xRefreshToken, err := models.GenerateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	appToken := xToken
	var appTokenExpiresIn int64 = int64(models.TokenExpiry.Seconds())

	if app := findApp(req.AppId); app != nil {
		remoteToken, remoteExpiresIn, remoteErr := requestAppToken(app, user.Sn, wxResp.OpenId, unionId, user.Id,
			user.NickName, user.Icon)
		if remoteErr != nil {
			LOGGER.Warn("remote app token failed, fallback to x-token", "appId", req.AppId, "err", remoteErr.Error())
		} else if remoteToken != "" {
			appToken = remoteToken
			appTokenExpiresIn = remoteExpiresIn
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": models.MiniLoginVO{
		Sn:            user.Sn,
		XToken:        xToken,
		XRefreshToken: xRefreshToken,
		AppToken:      appToken,
		TokenType:     "Bearer",
		ExpiresIn:     appTokenExpiresIn,
		User:          models.UserToVO(user),
	}})
}

func exchangeWxMiniCode(code string) (*WxMiniSessionResponse, error) {
	params := url.Values{}
	params.Set("appid", "wx_mini_appid")
	params.Set("secret", "wx_mini_secret")
	params.Set("js_code", code)
	params.Set("grant_type", "authorization_code")

	resp, err := restyClient.R().
		SetResult(&WxMiniSessionResponse{}).
		Get("https://api.weixin.qq.com/sns/jscode2session?" + params.Encode())

	if err != nil {
		return nil, err
	}

	return resp.Result().(*WxMiniSessionResponse), nil
}

func findApp(appId string) *internal.WxmTokenEntry {
	for i := range internal.Conf.Apps {
		if internal.Conf.Apps[i].Id == appId {
			return &internal.Conf.Apps[i]
		}
	}
	return nil
}

func requestAppToken(app *internal.WxmTokenEntry, sn, openId, unionId string, userId uint64, nickName, icon string) (string, int64, error) {
	cipher, err := models.EncryptAppPayload(app.Id, unionId)
	if err != nil {
		return "", 0, err
	}

	body := map[string]interface{}{
		"sn":        sn,
		"openId":    openId,
		"unionId":   unionId,
		"userId":    userId,
		"nickName":  nickName,
		"icon":      icon,
		"cipher":    cipher,
		"timestamp": time.Now().Unix(),
	}

	var result struct {
		Data AppTokenResponse `json:"data"`
	}

	resp, err := resty.New().SetTimeout(5 * time.Second).R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(&result).
		Post(app.TokenUrl)

	if err != nil {
		return "", 0, err
	}

	if resp.StatusCode() != 200 {
		LOGGER.Warn("app token response not 200", "status", resp.StatusCode(), "body", string(resp.Body()))
		return "", 0, fmt.Errorf("app token response status %d", resp.StatusCode())
	}

	return result.Data.Token, result.Data.ExpiresIn, nil
}
