package auth

import (
	"net/http"
	"github.com/zzznow/common"
	"time"

	"github.com/zzznow/z-uc/models"
	"github.com/gin-gonic/gin"

	"github.com/zzznow/z-uc/auth/internal"
)

type SmsLoginDTO struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type smsVerifyResp struct {
	Data struct {
		Verified bool   `json:"verified"`
		Phone    string `json:"phone"`
	} `json:"data"`
}

func SmsLogin(c *gin.Context) {
	var req SmsLoginDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := restyClient.R().
		SetBody(map[string]string{
			"phone": req.Phone,
			"code":  req.Code,
			"type":  "login",
		}).
		SetResult(&smsVerifyResp{}).
		Post(internal.Conf.BaseURL + "/sms/verify")
	if err != nil || resp.StatusCode() != http.StatusOK {
		common.ErrorMsg(c, http.StatusBadRequest, "z-uc-auth: verification code verification failed")
		return
	}

	result := resp.Result().(*smsVerifyResp)
	if !result.Data.Verified {
		common.ErrorMsg(c, http.StatusBadRequest, "z-uc-auth: verification code incorrect")
		return
	}

	loginName := req.Phone
	names, err := NamesRepo.GetByLoginName(loginName)
	if err != nil {
		nickName := models.RandomNickName()
		now := time.Now().UnixMilli()
		user := &models.User{
			Name:                  "",
			Password:              "",
			NickName:              nickName,
			Gender:                "N",
			CreateFrom:            "phone",
			Tel:                   req.Phone,
			CreateAt:              now,
			AccountNonExpired:     1,
			AccountNonLocked:      1,
			CredentialsNonExpired: 1,
			Enabled:               1,
		}

		tx, txErr := internal.Db.Beginx()
		if txErr != nil {
			common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: internal error")
			return
		}
		defer tx.Rollback()

		if !UserRepo.CreateTx(tx, user) {
			common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: internal error")
			return
		}
		user.Sn = models.GenerateSN(user.Id)
		tx.Exec("UPDATE t_user SET sn = ? WHERE id = ?", user.Sn, user.Id)

		n := &models.Names{
			LoginName: loginName,
			UserId:    user.Id,
			CreateAt:  now,
		}
		if !NamesRepo.CreateTx(tx, n) {
			common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: internal error")
			return
		}
		if cmtErr := tx.Commit(); cmtErr != nil {
			common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: internal error")
			return
		}

		respondWithTokens(c, user)
		return
	}

	user, err := UserRepo.GetById(names.UserId)
	if err != nil {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: user not found")
		return
	}

	respondWithTokens(c, user)
}
