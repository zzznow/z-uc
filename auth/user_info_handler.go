package auth

import (
	"net/http"
	"strconv"

	"github.com/zzznow/z-uc/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/zzznow/z-uc/auth/internal"
)

// ── 注册 ──────────────────────────────────────────────────

func Register(c *gin.Context) {
	var req models.SignUpDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var loginName string
	createFrom := req.Type

	switch req.Type {
	case "USERNAME":
		if req.Username == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username and password required"})
			return
		}
		loginName = req.Username
	case "EMAIL":
		if req.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email required"})
			return
		}
		loginName = req.Email
	case "TEL":
		if req.Tel == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tel required"})
			return
		}
		loginName = req.Tel
	case "WX_UNION":
		if req.WxUnionId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "wxUnionId required"})
			return
		}
		loginName = req.WxUnionId
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signup type"})
		return
	}

	if NamesRepo.Exists(loginName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account already exists"})
		return
	}

	passwordHash := ""
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "password encrypt failed"})
			return
		}
		passwordHash = string(hash)
	}

	now := NowMs()
	nickName := req.NickName
	if nickName == "" {
		nickName = models.RandomNickName()
	}

	user := &models.User{
		Password:              passwordHash,
		Name:                  req.Username,
		NickName:              nickName,
		Icon:                  req.Icon,
		Gender:                "N",
		CreateFrom:            createFrom,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "system error"})
		return
	}
	defer tx.Rollback()

	if !UserRepo.CreateTx(tx, user) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "register failed"})
		return
	}

	user.Sn = models.GenerateSN(user.Id)
	internal.Db.Exec("UPDATE t_user SET sn = ? WHERE id = ?", user.Sn, user.Id)

	names := &models.Names{
		LoginName: loginName,
		UserId:    user.Id,
		AppId:     "",
		CreateAt:  now,
	}
	if !NamesRepo.CreateTx(tx, names) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "register failed"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "register failed"})
		return
	}

	token, err := models.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models.RegisterVO{
		Id:        user.Id,
		Sn:        user.Sn,
		Username:  user.Name,
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: int64(models.TokenExpiry.Seconds()),
	}})
}

// ── 用户信息 CRUD ──────────────────────────────────────

func GetProfile(c *gin.Context) {
	sn := c.GetString("sn")
	user, err := UserRepo.GetBySn(sn)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": models.UserToVO(user)})
}

func UpdateProfile(c *gin.Context) {
	sn := c.GetString("sn")
	user, err := UserRepo.GetBySn(sn)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req models.ProfileUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.NickName != "" {
		user.NickName = req.NickName
	}
	if req.Icon != "" {
		user.Icon = req.Icon
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}
	if req.Birth != "" {
		user.Birth = req.Birth
	}
	if req.Location != "" {
		user.Location = req.Location
	}

	if !UserRepo.Update(user) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": models.UserToVO(user)})
}

func ChangePassword(c *gin.Context) {
	sn := c.GetString("sn")
	user, err := UserRepo.GetBySn(sn)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req models.PasswordChangeDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "old password incorrect"})
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password encrypt failed"})
		return
	}

	if !UserRepo.UpdatePassword(user.Id, string(newHash)) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password change failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password changed"})
}

func CancelAccount(c *gin.Context) {
	sn := c.GetString("sn")
	user, err := UserRepo.GetBySn(sn)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	NamesRepo.DeleteAllByUserId(user.Id)
	UserRepo.Delete(user.Id)
	c.JSON(http.StatusOK, gin.H{"message": "account cancelled"})
}

// ── 内部 API ────────────────────────────────────────────

func GetUserBySnInternal(c *gin.Context) {
	sn := c.Param("sn")
	user, err := UserRepo.GetBySn(sn)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": models.UserToVO(user)})
}

func GetUserByUnionIdInternal(c *gin.Context) {
	unionId := c.Param("unionId")
	user, err := UserRepo.GetByWxUnionId(unionId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": models.UserToVO(user)})
}

func GetUserIdInternal(c *gin.Context) {
	userIdStr := c.Query("userId")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}

	user, err := UserRepo.GetById(userId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": models.UserToVO(user)})
}

// ── Auth 中间�?─────────────────────────────────────────

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		tokenString := authHeader
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		claims, err := models.VerifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		sn, _ := claims["sn"].(string)
		c.Set("sn", sn)
		c.Set("claims", claims)
		c.Next()
	}
}
