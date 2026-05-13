package auth

import (
	LOGGER "log/slog"
	"net/http"

	"github.com/zzznow/z-uc/models"
	"github.com/gin-gonic/gin"
)

type GoogleTokenResponse struct {
	AccessToken string `json:"access_token"`
	IdToken     string `json:"id_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type GoogleUserInfo struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Locale        string `json:"locale"`
}

func GoogleToken(c *gin.Context) {
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

	redirectUri := req.RedirectUri
	if redirectUri == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "redirect_uri required"})
		return
	}

	googleResp, err := exchangeGoogleCode(req.Code, redirectUri)
	if err != nil {
		LOGGER.Error("google token exchange failed", "err", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "google auth failed"})
		return
	}

	userInfo, err := getGoogleUserInfo(googleResp.AccessToken)
	if err != nil {
		LOGGER.Error("google userinfo failed", "err", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "google auth failed"})
		return
	}

	if !userInfo.VerifiedEmail {
		c.JSON(http.StatusBadRequest, gin.H{"error": "google email not verified"})
		return
	}

	user, err := signUpOrLoginByThird("google", userInfo.Email, userInfo.Id, userInfo.Id+"-google",
		userInfo.Name, userInfo.Picture, userInfo.Email, userInfo.Locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	respondWithTokens(c, user)
}

func exchangeGoogleCode(code, redirectUri string) (*GoogleTokenResponse, error) {
	resp, err := restyClient.R().
		SetFormData(map[string]string{
			"code":          code,
			"client_id":     "639772744432-k42vv0st4j1960rh6qbobl2iat6srn4a.apps.googleusercontent.com",
			"client_secret": "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",
			"redirect_uri":  redirectUri,
			"grant_type":    "authorization_code",
		}).
		SetResult(&GoogleTokenResponse{}).
		Post("https://oauth2.googleapis.com/token")

	if err != nil {
		return nil, err
	}

	return resp.Result().(*GoogleTokenResponse), nil
}

func getGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	resp, err := restyClient.R().
		SetResult(&GoogleUserInfo{}).
		Get("https://openidconnect.googleapis.com/v1/userinfo?access_token=" + accessToken)

	if err != nil {
		return nil, err
	}

	return resp.Result().(*GoogleUserInfo), nil
}
