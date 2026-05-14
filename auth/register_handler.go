package auth

import (
	"net/http"
	"time"

	"github.com/zzznow/z-uc/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/zzznow/z-uc/auth/internal"
)

func SignUp(c *gin.Context) {
	var req models.SignUpDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var loginName string
	switch req.Type {
	case "email":
		if req.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email required"})
			return
		}
		loginName = req.Email
		if NamesRepo.Exists(loginName) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
	case "phone":
		if req.Tel == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone required"})
			return
		}
		loginName = req.Tel
		if NamesRepo.Exists(loginName) {
			c.JSON(http.StatusConflict, gin.H{"error": "phone already registered"})
			return
		}
	case "username":
		if req.Username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username required"})
			return
		}
		loginName = req.Username
		if NamesRepo.Exists(loginName) {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported register type"})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password required"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	now := time.Now().UnixMilli()

	nickName := req.NickName
	if nickName == "" {
		nickName = models.RandomNickName()
	}

	user := &models.User{
		Name:                  req.Username,
		Password:              string(hashed),
		NickName:              nickName,
		Icon:                  req.Icon,
		Gender:                "N",
		CreateFrom:            req.Type,
		Location:              req.Location,
		WxUnionId:             req.WxUnionId,
		Email:                 req.Email,
		Tel:                   req.Tel,
		CreateAt:              now,
		AccountNonExpired:     1,
		AccountNonLocked:      1,
		CredentialsNonExpired: 1,
		Enabled:               1,
	}

	tx, err := internal.Db.Beginx()
	if err != nil {
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

	names := &models.Names{
		LoginName: loginName,
		UserId:    user.Id,
		CreateAt:  now,
	}
	if !NamesRepo.CreateTx(tx, names) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	token, err := models.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}
	refreshToken, err := models.GenerateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models.LoginVO{
		Sn:           user.Sn,
		Token:        token,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(models.TokenExpiry.Seconds()),
		User:         models.UserToVO(user),
	}})
}