package auth

import (
	"net/http"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "verification code verification failed"})
		return
	}

	result := resp.Result().(*smsVerifyResp)
	if !result.Data.Verified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "verification code incorrect"})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		defer tx.Rollback()

		if !UserRepo.CreateTx(tx, user) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		if cmtErr := tx.Commit(); cmtErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		respondWithTokens(c, user)
		return
	}

	user, err := UserRepo.GetById(names.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	respondWithTokens(c, user)
}