package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"

	"github.com/siriusfreak/hack-zurich-2023/backend/internal/pallm"
)

type ChatMessage struct {
	ID       int    `json:"id"`
	ChatID   int    `json:"chat_id"`
	Message  string `json:"message"`
	IsBot    bool   `json:"is_bot"`
	Language string `json:"language"`
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

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/chat", func(c *gin.Context) {
		rows, err := db.Query("SELECT chat_id FROM chat_history GROUP BY chat_id")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
			return
		}
		defer rows.Close()

		var chatIDs []int
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
				return
			}
			chatIDs = append(chatIDs, id)
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "chatIDs": chatIDs})
	})

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
			return
		}

		resp, err := pallm.MakeRequest(msg.Message+"You should answer on language: "+msg.Language, pallm.RequestParameters{
			MaxOutputTokens: 2000,
			TopK:            20,
			TopP:            0.9,
		})

		if err != nil {
			c.JSON(200, gin.H{"status": err})
			return
		}

		_, err = db.Exec("INSERT INTO chat_history (chat_id, message, is_bot) VALUES (?, ?, ?)", chatID, msg.Message, msg.IsBot)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec("INSERT INTO chat_history (chat_id, message, is_bot) VALUES (?, ?, ?)", chatID, resp.Predictions[0].Content, true)
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(200, gin.H{"status": "message added", "response": resp.Predictions[0].Content})
	})

	r.Run()
}
