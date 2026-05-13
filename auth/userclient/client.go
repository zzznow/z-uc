// Package userclient �?替代 remote/client.go，直接调用函数，不走 HTTP
package userclient

import (
	"github.com/zzznow/z-uc/auth"
	"github.com/zzznow/z-uc/models"
)

type UserClient struct{}

func New(_ string) *UserClient {
	return &UserClient{}
}

func (c *UserClient) GetBySn(sn string) (*models.UserVO, error) {
	user, err := auth.UserRepo.GetBySn(sn)
	if err != nil {
		return nil, err
	}
	vo := models.UserToVO(user)
	return vo, nil
}

func (c *UserClient) GetById(userId uint64) (*models.UserVO, error) {
	user, err := auth.UserRepo.GetById(userId)
	if err != nil {
		return nil, err
	}
	vo := models.UserToVO(user)
	return vo, nil
}
