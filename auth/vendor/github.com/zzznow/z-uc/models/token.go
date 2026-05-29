package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func jwtSecret() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	return []byte("REAAW332LPPPPPEC00S++++++SDEDSSDFCCCCCC_____FFRFDSSDS")
}

const TokenExpiry = 30 * 24 * time.Hour
const RefreshTokenExpiry = 365 * 24 * time.Hour

func DeriveAppSecret(appId string) []byte {
	mac := hmac.New(sha256.New, jwtSecret())
	mac.Write([]byte("z-uc:app:" + appId))
	return mac.Sum(nil)
}

func EncryptAppPayload(appId, plaintext string) (string, error) {
	key := DeriveAppSecret(appId)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func DecryptAppPayload(appId, cipherHex string) (string, error) {
	key := DeriveAppSecret(appId)
	ciphertext, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func GenerateToken(user *User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":        "z-uc",
		"sub":        user.Sn,
		"iat":        now.Unix(),
		"exp":        now.Add(TokenExpiry).Unix(),
		"sn":         user.Sn,
		"name":       user.Name,
		"nickName":   user.NickName,
		"icon":       user.Icon,
		"tel":        user.Tel,
		"email":      user.Email,
		"gender":     user.Gender,
		"createFrom": user.CreateFrom,
		"location":   user.Location,
		"userId":     user.Id,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret())
}

func GenerateRefreshToken(user *User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":        "z-uc",
		"sub":        user.Sn,
		"iat":        now.Unix(),
		"exp":        now.Add(RefreshTokenExpiry).Unix(),
		"type":       "refresh",
		"sn":         user.Sn,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret())
}

func VerifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

func UserToVO(user *User) *UserVO {
	return &UserVO{
		Id:                   user.Id,
		Sn:                   user.Sn,
		Name:                 user.Name,
		NickName:             user.NickName,
		Icon:                 user.Icon,
		Gender:               user.Gender,
		Birth:                user.Birth,
		CreateFrom:           user.CreateFrom,
		Location:             user.Location,
		City:                 user.City,
		WxUnionId:            user.WxUnionId,
		Email:                user.Email,
		Tel:                  user.Tel,
		CreateAt:             user.CreateAt,
		AccountNonExpired:    user.AccountNonExpired,
		AccountNonLocked:     user.AccountNonLocked,
		CredentialsNonExpired: user.CredentialsNonExpired,
		Enabled:              user.Enabled,
	}
}
