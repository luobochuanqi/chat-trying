package connection

import (
	"chat/globals"
	"chat/utils"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func isUserExist(db *sql.DB, username string) bool {
	var count int
	if err := globals.QueryRowDb(db, "SELECT COUNT(*) FROM auth WHERE username = ?", username).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func ImportStudents(db *sql.DB) {
	csvPath := viper.GetString("student.csv")
	if csvPath == "" {
		globals.Warn("[csv] student csv path is empty, skipping import")
		return
	}

	globals.Info(fmt.Sprintf("[csv] looking for student csv at: %s", csvPath))

	if !utils.IsFileExist(csvPath) {
		globals.Warn(fmt.Sprintf("[csv] student csv file not found: %s", csvPath))
		return
	}

	initialMoney := viper.GetFloat64("student.initial_credit")
	initialDraws := viper.GetInt("student.initial_draws")
	if initialMoney <= 0 {
		initialMoney = 10.0
	}
	if initialDraws <= 0 {
		initialDraws = 50
	}

	file, err := os.Open(csvPath)
	if err != nil {
		globals.Warn(fmt.Sprintf("[csv] failed to open student csv: %s", err.Error()))
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		globals.Warn(fmt.Sprintf("[csv] failed to read student csv: %s", err.Error()))
		return
	}

	imported := 0
	for i, record := range records {
		if len(record) < 2 {
			continue
		}
		displayName := strings.TrimSpace(record[0])
		password := strings.TrimSpace(record[1])
		if displayName == "" || password == "" {
			continue
		}

		username := fmt.Sprintf("s%03d", i+1)

		if isUserExist(db, username) {
			continue
		}

		hashedPassword := utils.Sha2Encrypt(password)
		_, err := globals.ExecDb(db, `
			INSERT INTO auth (username, password, email, is_admin, bind_id, token, display_name)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, username, hashedPassword, fmt.Sprintf("%s@student.local", username), false, 0, "student", displayName)
		if err != nil {
			globals.Warn(fmt.Sprintf("[csv] failed to create user %s: %s", username, err.Error()))
			continue
		}

		var userId int64
		if err := globals.QueryRowDb(db, "SELECT id FROM auth WHERE username = ?", username).Scan(&userId); err != nil {
			continue
		}

		_, err = globals.ExecDb(db, `
			INSERT INTO quota (user_id, quota, used, credit_money, draw_count) VALUES (?, 0, 0, ?, ?)
		`, userId, initialMoney, initialDraws)
		if err != nil {
			globals.Warn(fmt.Sprintf("[csv] failed to create quota for %s: %s", username, err.Error()))
			continue
		}

		imported++
	}

	globals.Info(fmt.Sprintf("[csv] imported %d students from %s", imported, csvPath))
}
