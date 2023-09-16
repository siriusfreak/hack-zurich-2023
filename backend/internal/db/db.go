package db

import (
	"database/sql"
)

var DB *sql.DB

type ChatMessage struct {
	ID          int    `json:"id"`
	ChatID      int    `json:"chat_id"`
	Message     string `json:"message"`
	IsBot       bool   `json:"is_bot"`
	Language    string `json:"language"`
	RealMessage string `json:"real_message"`
}

func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return err
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS chat_history 
		(id INTEGER PRIMARY KEY, 
		chat_id INTEGER, 
		message TEXT, 
		is_bot BOOLEAN, 
		real_message TEXT)
	`)

	if err != nil {
		return err
	}

	return nil
}

type ChatInfo struct {
	ChatID int    `json:"chat_id"`
	Name   string `json:"name"`
}

func GetChatIDs() ([]ChatInfo, error) {
	rows, err := DB.Query(`
        SELECT 
            ch1.chat_id, 
            ch2.message AS name 
        FROM 
            (SELECT chat_id FROM chat_history GROUP BY chat_id) AS ch1 
        JOIN 
            chat_history AS ch2 
        ON 
            ch1.chat_id = ch2.chat_id 
        WHERE 
            ch2.id IN (SELECT MIN(id) FROM chat_history GROUP BY chat_id)
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chatInfos []ChatInfo
	for rows.Next() {
		var info ChatInfo
		if err := rows.Scan(&info.ChatID, &info.Name); err != nil {
			return nil, err
		}
		chatInfos = append(chatInfos, info)
	}

	return chatInfos, nil
}

func GetChatMessages(chatID int) ([]ChatMessage, error) {
	rows, err := DB.Query("SELECT id, chat_id, message, is_bot, real_message FROM chat_history WHERE chat_id = ?", chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.Message, &msg.IsBot, &msg.RealMessage); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func InsertChatMessage(chatID int, message string, realMessage string, isBot bool) error {
	_, err := DB.Exec("INSERT INTO chat_history (chat_id, message, real_message, is_bot) VALUES (?, ?, ?, ?)", chatID, message, realMessage, isBot)
	return err
}

func CloseDB() {
	DB.Close()
}
