package auth

import (
	"net/http"
	"github.com/zzznow/common"
	"strings"

	"github.com/zzznow/z-uc/models"
	"github.com/gin-gonic/gin"

	"github.com/zzznow/z-uc/auth/internal"
)

func VerifyTokenHandler(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		common.ErrorMsg(c, http.StatusUnauthorized, "z-uc-auth: unauthorized")
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := models.VerifyToken(tokenString)
	if err != nil {
		common.ErrorMsg(c, http.StatusUnauthorized, "z-uc-auth: invalid token")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models.TokenVerifyVO{
		Sn:         getStringClaim(claims, "sn"),
		Name:       getStringClaim(claims, "name"),
		NickName:   getStringClaim(claims, "nickName"),
		Icon:       getStringClaim(claims, "icon"),
		Email:      getStringClaim(claims, "email"),
		Tel:        getStringClaim(claims, "tel"),
		Gender:     getStringClaim(claims, "gender"),
		Location:   getStringClaim(claims, "location"),
		CreateFrom: getStringClaim(claims, "createFrom"),
	}})
}

func GetTokenInfo(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		common.ErrorMsg(c, http.StatusUnauthorized, "z-uc-auth: unauthorized")
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := models.VerifyToken(tokenString)
	if err != nil {
		common.ErrorMsg(c, http.StatusUnauthorized, "z-uc-auth: invalid token")
		return
	}

	sn, _ := claims["sn"].(string)
	var user models.User
	err = internal.Db.Get(&user, "SELECT * FROM t_user WHERE sn = ?", sn)
	if err != nil {
		common.ErrorMsg(c, http.StatusNotFound, "z-uc-auth: user not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models.UserToVO(&user)})
}

