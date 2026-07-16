package auth

import (
	"chat/channel"
	"chat/globals"
	"database/sql"
)

func (u *User) CreateInitialQuota(db *sql.DB) bool {
	_, err := globals.ExecDb(db, `
		INSERT INTO quota (user_id, quota, used) VALUES (?, ?, ?)
	`, u.GetID(db), channel.SystemInstance.GetInitialQuota(), 0.)
	return err == nil
}

func (u *User) GetQuota(db *sql.DB) float32 {
	var quota float32
	if err := globals.QueryRowDb(db, "SELECT quota FROM quota WHERE user_id = ?", u.GetID(db)).Scan(&quota); err != nil {
		return 0.
	}
	return quota
}

func (u *User) GetUsedQuota(db *sql.DB) float32 {
	var quota float32
	if err := globals.QueryRowDb(db, "SELECT used FROM quota WHERE user_id = ?", u.GetID(db)).Scan(&quota); err != nil {
		return 0.
	}
	return quota
}

func (u *User) SetQuota(db *sql.DB, quota float32) bool {
	_, err := globals.ExecDb(db, `
		INSERT INTO quota (user_id, quota, used) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE quota = ?
	`, u.GetID(db), quota, 0., quota)
	return err == nil
}

func (u *User) SetUsedQuota(db *sql.DB, used float32) bool {
	_, err := globals.ExecDb(db, `
		INSERT INTO quota (user_id, quota, used) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE used = ?
	`, u.GetID(db), 0., used, used)
	return err == nil
}

func (u *User) IncreaseQuota(db *sql.DB, quota float32) bool {
	_, err := globals.ExecDb(db, `
		INSERT INTO quota (user_id, quota, used) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE quota = quota + ?
	`, u.GetID(db), quota, 0., quota)
	return err == nil
}

func (u *User) IncreaseUsedQuota(db *sql.DB, used float32) bool {
	_, err := globals.ExecDb(db, `
		INSERT INTO quota (user_id, quota, used) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE used = used + ?
	`, u.GetID(db), 0., used, used)
	return err == nil
}

func (u *User) DecreaseQuota(db *sql.DB, quota float32) bool {
	_, err := globals.ExecDb(db, `
		INSERT INTO quota (user_id, quota, used) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE quota = quota - ?
	`, u.GetID(db), quota, 0., quota)
	return err == nil
}

func (u *User) UseQuota(db *sql.DB, quota float32) bool {
	if quota == 0 {
		return true
	}
	if !u.DecreaseQuota(db, quota) {
		return false
	}
	return u.IncreaseUsedQuota(db, quota)
}

func (u *User) PayedQuota(db *sql.DB, quota float32) bool {
	if quota == 0 {
		return true
	}

	current := u.GetQuota(db)
	if quota > current {
		return false
	}

	if !u.DecreaseQuota(db, quota) {
		return false
	}
	return u.IncreaseUsedQuota(db, quota)
}

func (u *User) PayedQuotaAsAmount(db *sql.DB, amount float32) bool {
	return u.PayedQuota(db, amount*10)
}

func (u *User) GetDrawCount(db *sql.DB) int {
	var count int
	if err := globals.QueryRowDb(db, "SELECT draw_count FROM quota WHERE user_id = ?", u.GetID(db)).Scan(&count); err != nil {
		return 0
	}
	return count
}

func (u *User) DecreaseDrawCount(db *sql.DB, n int) bool {
	_, err := globals.ExecDb(db, `
		UPDATE quota SET draw_count = draw_count - ? WHERE user_id = ?
	`, n, u.GetID(db))
	return err == nil
}

func (u *User) GetCreditMoney(db *sql.DB) float32 {
	var money float32
	if err := globals.QueryRowDb(db, "SELECT credit_money FROM quota WHERE user_id = ?", u.GetID(db)).Scan(&money); err != nil {
		return 0.
	}
	return money
}

func (u *User) SetCreditMoney(db *sql.DB, money float32) bool {
	_, err := globals.ExecDb(db, `
		INSERT INTO quota (user_id, credit_money) VALUES (?, ?) ON DUPLICATE KEY UPDATE credit_money = ?
	`, u.GetID(db), money, money)
	return err == nil
}

func (u *User) SetDrawCount(db *sql.DB, count int) bool {
	_, err := globals.ExecDb(db, `
		INSERT INTO quota (user_id, draw_count) VALUES (?, ?) ON DUPLICATE KEY UPDATE draw_count = ?
	`, u.GetID(db), count, count)
	return err == nil
}

func (u *User) CreateInitialQuotaWithDraw(db *sql.DB, money float32, draws int) bool {
	_, err := globals.ExecDb(db, `
		INSERT INTO quota (user_id, quota, used, credit_money, draw_count) VALUES (?, ?, 0, ?, ?)
	`, u.GetID(db), money, money, draws)
	return err == nil
}
