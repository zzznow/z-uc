package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	LOGGER "log/slog"
	"net/http"
	"time"

	"github.com/zzznow/z-uc/models"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"

	"github.com/zzznow/z-uc/auth/internal"
)

var restyClient = resty.New().SetTimeout(10 * time.Second)

func signUpOrLoginByThird(createFrom, loginName, wxUnionId, name, nickName, icon, email, location string) (*models.User, error) {
	var names models.Names
	err := internal.Db.Get(&names, "SELECT login_name, user_id, app_id, create_at FROM t_names WHERE login_name = ?", loginName)
	if err != nil {
		return createThirdUser(createFrom, loginName, wxUnionId, name, nickName, icon, email, location)
	}

	var user models.User
	err = internal.Db.Get(&user, "SELECT * FROM t_user WHERE id = ?", names.UserId)
	if err != nil {
		return createThirdUser(createFrom, loginName, wxUnionId, name, nickName, icon, email, location)
	}

	needUpdate := false
	if nickName != "" && user.NickName != nickName {
		user.NickName = nickName
		needUpdate = true
	}
	if icon != "" && user.Icon != icon {
		user.Icon = icon
		needUpdate = true
	}
	if email != "" && user.Email != email {
		user.Email = email
		needUpdate = true
	}
	if location != "" && user.Location != location {
		user.Location = location
		needUpdate = true
	}

	if needUpdate {
		internal.Db.Exec("UPDATE t_user SET nick_name=?, icon=?, email=?, location=? WHERE id=?",
			user.NickName, user.Icon, user.Email, user.Location, user.Id)
	}

	return &user, nil
}

func createThirdUser(createFrom, loginName, wxUnionId, name, nickName, icon, email, location string) (*models.User, error) {
	tx, err := internal.Db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if nickName == "" {
		nickName = models.RandomNickName()
	}

	now := time.Now().UnixMilli()

	user := &models.User{
		Name:                  name,
		Password:              "",
		NickName:              nickName,
		Icon:                  icon,
		Gender:                "N",
		CreateFrom:            createFrom,
		Location:              location,
		WxUnionId:             wxUnionId,
		Email:                 email,
		CreateAt:              now,
		AccountNonExpired:     1,
		AccountNonLocked:      1,
		CredentialsNonExpired: 1,
		Enabled:               1,
	}

	sql := `INSERT INTO t_user (sn, name, password, nick_name, icon, gender, birth, create_from, location, city, wx_union_id, email, tel, create_at, account_non_expired, account_non_locked, credentials_non_expired, enabled) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := tx.Exec(sql,
		user.Sn, user.Name, user.Password, user.NickName, user.Icon, user.Gender,
		user.Birth, user.CreateFrom, user.Location, user.City, user.WxUnionId,
		user.Email, user.Tel, user.CreateAt, user.AccountNonExpired, user.AccountNonLocked,
		user.CredentialsNonExpired, user.Enabled)
	if err != nil {
		LOGGER.Error("create third user failed", "err", err.Error())
		return nil, err
	}

	id, _ := result.LastInsertId()
	user.Id = uint64(id)
	user.Sn = models.GenerateSN(user.Id)

	tx.Exec("UPDATE t_user SET sn = ? WHERE id = ?", user.Sn, user.Id)

	_, err = tx.Exec("INSERT INTO t_names (login_name, user_id, app_id, create_at) VALUES (?, ?, ?, ?)",
		loginName, user.Id, "", now)
	if err != nil {
		LOGGER.Error("create names failed", "err", err.Error())
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func respondWithTokens(c *gin.Context, user *models.User) {
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

func getStringClaim(claims map[string]interface{}, key string) string {
	if v, ok := claims[key].(string); ok {
		return v
	}
	return ""
}

func NewState(c *gin.Context) {
	b := make([]byte, 16)
	rand.Read(b)
	state := hex.EncodeToString(b)

	if internal.RDB != nil {
		internal.RDB.Set(context.Background(), "oauth_state:"+state, "1", 5*time.Minute)
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"state": state}})
}

func saveState(c *gin.Context, state string) error {
	if internal.RDB != nil {
		return internal.RDB.Set(c.Request.Context(), "oauth_state:"+state, "1", 5*time.Minute).Err()
	}
	return nil
}

func getState(c *gin.Context, state string) (string, error) {
	if state == "" {
		return "", nil
	}
	if internal.RDB != nil {
		val, err := internal.RDB.GetDel(c.Request.Context(), "oauth_state:"+state).Result()
		if err != nil {
			return "", err
		}
		return val, nil
	}
	return state, nil
}
