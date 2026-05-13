package models

type Names struct {
	LoginName string `db:"login_name"`
	UserId    uint64 `db:"user_id"`
	AppId     string `db:"app_id"`
	CreateAt  int64  `db:"create_at"`
}

func (Names) TableName() string {
	return "t_names"
}
