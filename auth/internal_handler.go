package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zzznow/common"
	"github.com/zzznow/z-uc/auth/internal"
	"github.com/zzznow/z-uc/models"
	"golang.org/x/crypto/bcrypt"
)

// ── 内部 API：供 gaming-partner 等下游服务调用 ────────────────
// 不需要用户 z-uc Token，通过 internal secret 或内网鉴权即可访问

// InternalBindPhone 绑定手机号（内部调用）
// POST /internal/auth/bind-phone  { sn, phone, code }
func InternalBindPhone(c *gin.Context) {
	var req struct {
		Sn    string `json:"sn" binding:"required"`
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := UserRepo.GetBySn(req.Sn)
	if err != nil {
		common.ErrorMsg(c, http.StatusNotFound, "z-uc-auth: user not found")
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

	// 删除该用户旧的 PHONE 关联
	if user.Tel != "" {
		NamesRepo.DeleteByLoginName(user.Tel)
	}

	// 更新 t_user
	if !UserRepo.UpdatePhone(user.Id, req.Phone) {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: bind phone failed")
		return
	}

	// 写入 t_names
	now := NowMs()
	_, err = internal.Db.Exec("INSERT INTO t_names (login_name, user_id, app_id, create_at) VALUES (?, ?, ?, ?)",
		req.Phone, user.Id, "", now)
	if err != nil {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: bind phone failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "phone bound"})
}

// InternalBindUsername 绑定用户名（内部调用）
// POST /internal/auth/bind-username  { sn, username }
func InternalBindUsername(c *gin.Context) {
	var req struct {
		Sn       string `json:"sn" binding:"required"`
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := UserRepo.GetBySn(req.Sn)
	if err != nil {
		common.ErrorMsg(c, http.StatusNotFound, "z-uc-auth: user not found")
		return
	}

	// 检查用户名是否已被占用
	if NamesRepo.Exists(req.Username) {
		common.ErrorMsg(c, http.StatusBadRequest, "z-uc-auth: username already taken")
		return
	}

	// 删除该用户旧的 USERNAME 关联
	if user.Name != "" {
		NamesRepo.DeleteByLoginName(user.Name)
	}

	// 更新 t_user
	if !UserRepo.UpdateName(user.Id, req.Username) {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: bind username failed")
		return
	}

	// 写入 t_names
	now := NowMs()
	_, err = internal.Db.Exec("INSERT INTO t_names (login_name, user_id, app_id, create_at) VALUES (?, ?, ?, ?)",
		req.Username, user.Id, "", now)
	if err != nil {
		common.ErrorMsg(c, http.StatusInternalServerError, "z-uc-auth: bind username failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "username bound"})
}

// InternalChangePassword 修改密码（内部调用）
// POST /internal/auth/change-password  { sn, oldPassword, newPassword }
func InternalChangePassword(c *gin.Context) {
	var req struct {
		Sn          string `json:"sn" binding:"required"`
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := UserRepo.GetBySn(req.Sn)
	if err != nil {
		common.ErrorMsg(c, http.StatusNotFound, "z-uc-auth: user not found")
		return
	}

	// 验证旧密码
	if user.Password != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
			common.ErrorMsg(c, http.StatusUnauthorized, "z-uc-auth: old password incorrect")
			return
		}
	}

	// 生成新密码哈希
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

// InternalGetProfile 获取用户信息（内部调用，通过 sn）
// GET /internal/auth/profile/:sn
func InternalGetProfile(c *gin.Context) {
	sn := c.Param("sn")
	user, err := UserRepo.GetBySn(sn)
	if err != nil {
		common.ErrorMsg(c, http.StatusNotFound, "z-uc-auth: user not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": models.UserToVO(user)})
}

// InternalVerifyToken 验证 z-uc Token 并返回用户信息（内部调用）
// POST /internal/auth/verify-token  { token }
func InternalVerifyToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	tokenString := req.Token
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	claims, err := models.VerifyToken(tokenString)
	if err != nil {
		common.ErrorMsg(c, http.StatusUnauthorized, "z-uc-auth: invalid token")
		return
	}

	sn, _ := claims["sn"].(string)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"sn": sn, "claims": claims}})
}
