package auth

import (
	LOGGER "log/slog"
	"time"

	"github.com/fndome/xb"
	"github.com/zzznow/z-uc/models"
	"github.com/jmoiron/sqlx"

	"github.com/zzznow/z-uc/auth/internal"
)

// ── UserRepository ────────────────────────────────────────

type UserRepository struct{}

var UserRepo = &UserRepository{}

func (r *UserRepository) GetById(id uint64) (*models.User, error) {
	var user models.User
	sql, vs, _ := xb.Of(&user).Eq("id", id).Build().SqlOfSelect()
	err := internal.Db.Get(&user, sql, vs...)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetBySn(sn string) (*models.User, error) {
	var user models.User
	sql, vs, _ := xb.Of(&user).Eq("sn", sn).Build().SqlOfSelect()
	err := internal.Db.Get(&user, sql, vs...)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByName(name string) (*models.User, error) {
	var user models.User
	sql, vs, _ := xb.Of(&user).Eq("name", name).Build().SqlOfSelect()
	err := internal.Db.Get(&user, sql, vs...)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByWxUnionId(wxUnionId string) (*models.User, error) {
	var user models.User
	sql, vs, _ := xb.Of(&user).Eq("wx_union_id", wxUnionId).Build().SqlOfSelect()
	err := internal.Db.Get(&user, sql, vs...)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	sql, vs, _ := xb.Of(&user).Eq("email", email).Build().SqlOfSelect()
	err := internal.Db.Get(&user, sql, vs...)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CreateTx(tx *sqlx.Tx, user *models.User) bool {
	sql, vs := xb.Of(user).
		Insert(func(b *xb.InsertBuilder) {
			b.Set("sn", user.Sn).
				Set("name", user.Name).
				Set("password", user.Password).
				Set("nick_name", user.NickName).
				Set("icon", user.Icon).
				Set("gender", user.Gender).
				Set("birth", user.Birth).
				Set("create_from", user.CreateFrom).
				Set("location", user.Location).
				Set("city", user.City).
				Set("wx_union_id", user.WxUnionId).
				Set("email", user.Email).
				Set("tel", user.Tel).
				Set("create_at", user.CreateAt).
				Set("account_non_expired", user.AccountNonExpired).
				Set("account_non_locked", user.AccountNonLocked).
				Set("credentials_non_expired", user.CredentialsNonExpired).
				Set("enabled", user.Enabled)
		}).
		Build().
		SqlOfInsert()

	result, err := tx.Exec(sql, vs...)
	if err != nil {
		LOGGER.Error("[UserRepository.CreateTx]", "err", err.Error())
		return false
	}
	id, _ := result.LastInsertId()
	user.Id = uint64(id)
	return true
}

func (r *UserRepository) Update(user *models.User) bool {
	sql, vs := xb.Of(user).
		Update(func(ub *xb.UpdateBuilder) {
			ub.Set("nick_name", user.NickName).
				Set("icon", user.Icon).
				Set("gender", user.Gender).
				Set("birth", user.Birth).
				Set("location", user.Location).
				Set("city", user.City)
		}).
		Eq("id", user.Id).
		Build().
		SqlOfUpdate()

	_, err := internal.Db.Exec(sql, vs...)
	if err != nil {
		LOGGER.Error("[UserRepository.Update]", "err", err.Error())
		return false
	}
	return true
}

func (r *UserRepository) UpdatePassword(userId uint64, newPassword string) bool {
	var user models.User
	sql, vs := xb.Of(&user).
		Update(func(ub *xb.UpdateBuilder) {
			ub.Set("password", newPassword)
		}).
		Eq("id", userId).
		Build().
		SqlOfUpdate()

	_, err := internal.Db.Exec(sql, vs...)
	if err != nil {
		LOGGER.Error("[UserRepository.UpdatePassword]", "err", err.Error())
		return false
	}
	return true
}

func (r *UserRepository) Delete(id uint64) bool {
	_, err := internal.Db.Exec("DELETE FROM t_user WHERE id = ?", id)
	if err != nil {
		LOGGER.Error("[UserRepository.Delete]", "err", err.Error())
		return false
	}
	return true
}

// ── NamesRepository ───────────────────────────────────────

type NamesRepository struct{}

var NamesRepo = &NamesRepository{}

func (r *NamesRepository) Exists(loginName string) bool {
	var count int
	err := internal.Db.Get(&count, "SELECT COUNT(*) FROM t_names WHERE login_name = ?", loginName)
	if err != nil {
		return false
	}
	return count > 0
}

func (r *NamesRepository) GetByLoginName(loginName string) (*models.Names, error) {
	var names models.Names
	sql, vs, _ := xb.Of(&names).Eq("login_name", loginName).Build().SqlOfSelect()
	err := internal.Db.Get(&names, sql, vs...)
	if err != nil {
		return nil, err
	}
	return &names, nil
}

func (r *NamesRepository) CreateTx(tx *sqlx.Tx, names *models.Names) bool {
	sql, vs := xb.Of(names).
		Insert(func(b *xb.InsertBuilder) {
			b.Set("login_name", names.LoginName).
				Set("user_id", names.UserId).
				Set("app_id", names.AppId).
				Set("create_at", names.CreateAt)
		}).
		Build().
		SqlOfInsert()

	_, err := tx.Exec(sql, vs...)
	if err != nil {
		LOGGER.Error("[NamesRepository.CreateTx]", "err", err.Error())
		return false
	}
	return true
}

func (r *NamesRepository) DeleteByLoginName(loginName string) bool {
	_, err := internal.Db.Exec("DELETE FROM t_names WHERE login_name = ?", loginName)
	if err != nil {
		LOGGER.Error("[NamesRepository.DeleteByLoginName]", "err", err.Error())
		return false
	}
	return true
}

func (r *NamesRepository) DeleteAllByUserId(userId uint64) bool {
	_, err := internal.Db.Exec("DELETE FROM t_names WHERE user_id = ?", userId)
	if err != nil {
		LOGGER.Error("[NamesRepository.DeleteAllByUserId]", "err", err.Error())
		return false
	}
	return true
}

func NowMs() int64 {
	return time.Now().UnixMilli()
}
