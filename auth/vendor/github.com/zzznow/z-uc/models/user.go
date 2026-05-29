package models

type User struct {
	Id                    uint64 `db:"id"`
	Sn                    string `db:"sn"`
	Name                  string `db:"name"`
	Password              string `db:"password"`
	NickName              string `db:"nick_name"`
	Icon                  string `db:"icon"`
	Gender                string `db:"gender"`
	Birth                 string `db:"birth"`
	CreateFrom            string `db:"create_from"`
	Location              string `db:"location"`
	City                  string `db:"city"`
	WxUnionId             string `db:"wx_union_id"`
	Email                 string `db:"email"`
	Tel                   string `db:"tel"`
	CreateAt              int64  `db:"create_at"`
	AccountNonExpired     int    `db:"account_non_expired"`
	AccountNonLocked      int    `db:"account_non_locked"`
	CredentialsNonExpired int    `db:"credentials_non_expired"`
	Enabled               int    `db:"enabled"`
}

func (User) TableName() string {
	return "t_user"
}
