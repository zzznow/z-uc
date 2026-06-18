package auth

import (
	LOGGER "log/slog"
	"net/http"
	"net/url"

	"github.com/zzznow/common"
	"github.com/zzznow/z-uc/models"
	"github.com/gin-gonic/gin"
)

type DouyinSessionResponse struct {
	ErrCode    int    `json:"err_no"`
	ErrMsg     string `json:"err_tips"`
	OpenId     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionId    string `json:"unionid"`
}

func DouyinToken(c *gin.Context) {
	var req models.ThirdLoginDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	dyResp, err := exchangeDouyinCode(req.Code)
	if err != nil {
		LOGGER.Error("douyin code exchange failed", "err", err.Error())
		common.ErrorMsg(c, http.StatusBadRequest, "z-uc-auth: douyin auth failed")
		return
	}

	if dyResp.ErrCode != 0 {
		common.ErrorMsg(c, http.StatusBadRequest, dyResp.ErrMsg)
		return
	}

	openId := dyResp.OpenId
	if openId == "" {
		common.ErrorMsg(c, http.StatusBadRequest, "z-uc-auth: empty openid")
		return
	}

	user, err := signUpOrLoginByThird("dy", openId, openId, openId+"-dy",
		"抖音用户"+openId[len(openId)-4:], "", "", "")
	if err != nil {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: login failed")
		return
	}

	respondWithTokens(c, user)
}

func exchangeDouyinCode(code string) (*DouyinSessionResponse, error) {
	params := url.Values{}
	params.Set("appid", "tt_appid")
	params.Set("secret", "tt_secret")
	params.Set("code", code)

	resp, err := restyClient.R().
		SetResult(&DouyinSessionResponse{}).
		Get("https://developer.toutiao.com/api/apps/v2/jscode2session?" + params.Encode())

	if err != nil {
		return nil, err
	}

	return resp.Result().(*DouyinSessionResponse), nil
}
