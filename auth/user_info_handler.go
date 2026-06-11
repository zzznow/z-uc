package auth

import (
	"net/http"
	"strconv"

	"github.com/zzznow/common"

	"github.com/gin-gonic/gin"
	"github.com/zzznow/z-uc/models"
	"golang.org/x/crypto/bcrypt"

	"github.com/zzznow/z-uc/auth/internal"
)

// ── 注册 ──────────────────────────────────────────────────

func Register(c *gin.Context) {
	var req models.SignUpDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	var loginName string
	createFrom := req.Type

	switch req.Type {
	case "USERNAME":
		if req.Username == "" || req.Password == "" {
			common.ErrorMsg(c, http.StatusBadRequest, "z-uc-auth: username and password required")
			return
		}
		loginName = req.Username
	case "EMAIL":
		if req.Email == "" {
			common.ErrorMsg(c, http.StatusBadRequest, "z-uc-auth: email required")
			return
		}
		loginName = req.Email
	case "TEL":
		if req.Tel == "" {
			common.ErrorMsg(c, http.StatusBadRequest, "z-uc-auth: tel required")
			return
		}
		loginName = req.Tel
	case "WX_UNION":
		if req.WxUnionId == "" {
			common.ErrorMsg(c, http.StatusBadRequest, "z-uc-auth: wxUnionId required")
			return
		}
		loginName = req.WxUnionId
	default:
		common.ErrorMsg(c, http.StatusBadRequest, "z-uc-auth: invalid signup type")
		return
	}

	if NamesRepo.Exists(loginName) {
		common.ErrorMsg(c, http.StatusBadRequest, "z-uc-auth: account already exists")
		return
	}

	passwordHash := ""
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: password encrypt failed")
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
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: system error")
		return
	}
	defer tx.Rollback()

	if !UserRepo.CreateTx(tx, user) {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: register failed")
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
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: register failed")
		return
	}

	if err := tx.Commit(); err != nil {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: register failed")
		return
	}

	token, err := models.GenerateToken(user)
	if err != nil {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: token generation failed")
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
		common.ErrorMsg(c, http.StatusNotFound, "z-uc-auth: user not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": models.UserToVO(user)})
}

func UpdateProfile(c *gin.Context) {
	sn := c.GetString("sn")
	user, err := UserRepo.GetBySn(sn)
	if err != nil {
		common.ErrorMsg(c, http.StatusNotFound, "z-uc-auth: user not found")
		return
	}

	var req models.ProfileUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorMsg(c, http.StatusBadRequest, err.Error())
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
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: update failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": models.UserToVO(user)})
}

func ChangePassword(c *gin.Context) {
	sn := c.GetString("sn")
	user, err := UserRepo.GetBySn(sn)
	if err != nil {
		common.ErrorMsg(c, http.StatusNotFound, "z-uc-auth: user not found")
		return
	}

	var req models.PasswordChangeDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		common.ErrorMsg(c, http.StatusUnauthorized, "z-uc-auth: old password incorrect")
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: password encrypt failed")
		return
	}

	if !UserRepo.UpdatePassword(user.Id, string(newHash)) {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: password change failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password changed"})
}

// ── 绑定手机号 ──────────────────────────────────────────

func BindPhone(c *gin.Context) {
	sn := c.GetString("sn")
	user, err := UserRepo.GetBySn(sn)
	if err != nil {
		common.ErrorMsg(c, http.StatusNotFound, "z-uc-auth: user not found")
		return
	}

	var req models.BindPhoneDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	// TODO: 验证 SMS 验证码（接入 z-3sp 内部验证接口后补全）

	// 检查手机号是否已被其他用户占用
	if NamesRepo.Exists(req.Phone) {
		existing, _ := NamesRepo.GetByLoginName(req.Phone)
		if existing != nil && existing.UserId != user.Id {
			common.ErrorMsg(c, http.StatusBadRequest, "z-uc-auth: phone already bound to another account")
			return
		}
	}

	// 删除该用户旧的 PHONE 类型关联（如果有）
	oldNames, _ := NamesRepo.GetByUserId(user.Id)
	for _, n := range oldNames {
		if n.LoginName == user.Tel {
			NamesRepo.DeleteByLoginName(n.LoginName)
		}
	}

	// 更新 t_user 表
	if !UserRepo.UpdatePhone(user.Id, req.Phone) {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: bind phone failed")
		return
	}

	// 写入 t_names 表
	now := NowMs()
	_, err = internal.Db.Exec("INSERT INTO t_names (login_name, user_id, app_id, create_at) VALUES (?, ?, ?, ?)",
		req.Phone, user.Id, "", now)
	if err != nil {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: bind phone failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "phone bound"})
}

// ── 绑定用户名 ──────────────────────────────────────────

func BindUsername(c *gin.Context) {
	sn := c.GetString("sn")
	user, err := UserRepo.GetBySn(sn)
	if err != nil {
		common.ErrorMsg(c, http.StatusNotFound, "z-uc-auth: user not found")
		return
	}

	var req models.BindUsernameDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	// 检查用户名是否已被占用
	if NamesRepo.Exists(req.Username) {
		common.ErrorMsg(c, http.StatusBadRequest, "z-uc-auth: username already taken")
		return
	}

	// 删除该用户旧的 USERNAME 类型关联
	if user.Name != "" {
		NamesRepo.DeleteByLoginName(user.Name)
	}

	// 更新 t_user 表
	if !UserRepo.UpdateName(user.Id, req.Username) {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: bind username failed")
		return
	}

	// 写入 t_names 表
	now := NowMs()
	_, err = internal.Db.Exec("INSERT INTO t_names (login_name, user_id, app_id, create_at) VALUES (?, ?, ?, ?)",
		req.Username, user.Id, "", now)
	if err != nil {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: bind username failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "username bound"})
}

// ── 注销账号 ──────────────────────────────────────────

func CancelAccount(c *gin.Context) {
	sn := c.GetString("sn")
	user, err := UserRepo.GetBySn(sn)
	if err != nil {
		common.ErrorMsg(c, http.StatusNotFound, "z-uc-auth: user not found")
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
		common.ErrorMsg(c, http.StatusNotFound, "z-uc-auth: user not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": models.UserToVO(user)})
}

func GetUserByUnionIdInternal(c *gin.Context) {
	unionId := c.Param("unionId")
	user, err := UserRepo.GetByWxUnionId(unionId)
	if err != nil {
		common.ErrorMsg(c, http.StatusNotFound, "z-uc-auth: user not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": models.UserToVO(user)})
}

func GetUserIdInternal(c *gin.Context) {
	userIdStr := c.Query("userId")
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		common.ErrorMsg(c, http.StatusBadRequest, "z-uc-auth: invalid userId")
		return
	}

	user, err := UserRepo.GetById(userId)
	if err != nil {
		common.ErrorMsg(c, http.StatusNotFound, "z-uc-auth: user not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": models.UserToVO(user)})
}

// -- Auth Middleware -------------------------------------------------

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.ErrorMsg(c, http.StatusUnauthorized, "z-uc-auth: unauthorized")
			c.Abort()
			return
		}

		tokenString := authHeader
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		claims, err := models.VerifyToken(tokenString)
		if err != nil {
			common.ErrorMsg(c, http.StatusUnauthorized, "z-uc-auth: invalid token")
			c.Abort()
			return
		}

		sn, _ := claims["sn"].(string)
		c.Set("sn", sn)
		c.Set("claims", claims)
		c.Next()
	}
}
