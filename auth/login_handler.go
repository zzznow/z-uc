package auth

import (
	"net/http"

	"github.com/zzznow/z-uc/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/zzznow/z-uc/auth/internal"
)

func FormLogin(c *gin.Context) {
	var req models.LoginDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var names models.Names
	err := internal.Db.Get(&names, "SELECT login_name, user_id, app_id, create_at FROM t_names WHERE login_name = ?", req.LoginName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "username or password error"})
		return
	}

	var user models.User
	err = internal.Db.Get(&user, "SELECT * FROM t_user WHERE id = ?", names.UserId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "username or password error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "username or password error"})
		return
	}

	if user.Enabled != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "account disabled"})
		return
	}

	token, err := models.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	refreshToken, err := models.GenerateRefreshToken(&user)
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
		User:         models.UserToVO(&user),
	}})
}

func RefreshToken(c *gin.Context) {
	var req models.TokenRefreshDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := models.VerifyToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not a refresh token"})
		return
	}

	sn, _ := claims["sn"].(string)
	var user models.User
	err = internal.Db.Get(&user, "SELECT * FROM t_user WHERE sn = ?", sn)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	token, err := models.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	refreshToken, err := models.GenerateRefreshToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models.TokenRefreshVO{
		Sn:           user.Sn,
		Token:        token,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(models.TokenExpiry.Seconds()),
	}})
}
