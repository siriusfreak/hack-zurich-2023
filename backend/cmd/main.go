package main

import (
	"database/sql"
	"github.com/siriusfreak/hack-zurich-2023/backend/internal/chatgpt"
	"github.com/siriusfreak/hack-zurich-2023/backend/internal/elastic"
	"github.com/siriusfreak/hack-zurich-2023/backend/internal/embeddings"
	"github.com/siriusfreak/hack-zurich-2023/backend/internal/templater"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
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

	tmpl, err := templater.New("backend/config/templates.yaml")
	if err != nil {
		log.Fatal(err)
	}

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
			c.JSON(500, gin.H{"status": err})
			return
		}

		embed, err := embeddings.MakePredictionRequest("hackzurich23-8200", embeddings.PredictRequest{
			Instances: []embeddings.Instance{
				{
					Text: msg.Message,
				},
			},
		})

		if err != nil {
			c.JSON(500, gin.H{"status": err})
			return
		}

		request := elastic.SearchRequest{}
		request.KNN.Field = "embedding"
		request.KNN.QueryVector = embed.Predictions[0].TextEmbedding
		request.KNN.K = 10
		request.KNN.NumCandidates = 10
		request.Size = 10

		searchResp, err := elastic.Search("sika_chat_index", request)

		documents := make([]templater.Document, 0, len(searchResp.Hits.Hits))
		for _, hit := range searchResp.Hits.Hits {
			documents = append(documents, templater.Document{
				Url:     hit.Source.Links[0],
				Offset:  hit.Source.Offset,
				Content: hit.Source.Content,
			})
		}

		if err != nil {
			c.JSON(500, gin.H{"status": err})
		}

		template, err := tmpl.ProcessTemplateInitQuestionData([]templater.InitQuestionData{
			{
				Language:  msg.Language,
				Question:  msg.Message,
				Documents: documents,
			},
		})

		if err != nil {
			c.JSON(200, gin.H{"status": err})
			return
		}

		resp, err := chatgpt.CallAPI(chatgpt.RequestBody{
			Model: "gpt-3.5-turbo",
			Messages: []chatgpt.Message{
				{
					Role:    "user",
					Content: template,
				},
			},
		})

		if err != nil {
			c.JSON(200, gin.H{"status": err})
			return
		}

		_, err = db.Exec("INSERT INTO chat_history (chat_id, message, is_bot) VALUES (?, ?, ?)", chatID, msg.Message, msg.IsBot)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec("INSERT INTO chat_history (chat_id, message, is_bot) VALUES (?, ?, ?)", chatID, resp.Choices[0].Message.Content, true)
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(200, gin.H{"status": "message added", "response": resp.Choices[0].Message.Content})
	})

	r.Run()
}
