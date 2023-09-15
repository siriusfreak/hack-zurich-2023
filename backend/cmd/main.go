package main

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type ChatMessage struct {
	ID      int    `json:"id"`
	ChatID  int    `json:"chat_id"`
	Message string `json:"message"`
	IsBot   bool   `json:"is_bot"`
}

func main() {
	db, err := sql.Open("sqlite3", "chat.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS chat_history (id INTEGER PRIMARY KEY, chat_id INTEGER, message TEXT, is_bot BOOLEAN)`)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	r.GET("/chat/:chatID", func(c *gin.Context) {
		chatID := c.Param("chatID")
		rows, err := db.Query("SELECT id, chat_id, message, is_bot FROM chat_history WHERE chat_id = ?", chatID)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var messages []ChatMessage
		for rows.Next() {
			var msg ChatMessage
			if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.Message, &msg.IsBot); err != nil {
				log.Fatal(err)
			}
			messages = append(messages, msg)
		}

		c.JSON(200, messages)
	})

	r.POST("/chat/:chatID", func(c *gin.Context) {
		chatID := c.Param("chatID")
		var msg ChatMessage
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		msg.ChatID, err = strconv.Atoi(chatID)
		if err != nil {
			c.JSON(200, gin.H{"status": err})
		}
		_, err = db.Exec("INSERT INTO chat_history (chat_id, message, is_bot) VALUES (?, ?, ?)", chatID, msg.Message, msg.IsBot)
		if err != nil {
			log.Fatal(err)
		}
		_, err = db.Exec("INSERT INTO chat_history (chat_id, message, is_bot) VALUES (?, ?, ?)", chatID, "accepted", true)
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(200, gin.H{"status": "message added"})
	})

	r.Run()
}
