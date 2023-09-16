package main

import (
	"fmt"
	"github.com/siriusfreak/hack-zurich-2023/backend/internal/chatgpt"
	"github.com/siriusfreak/hack-zurich-2023/backend/internal/db"
	"github.com/siriusfreak/hack-zurich-2023/backend/internal/elastic"
	"github.com/siriusfreak/hack-zurich-2023/backend/internal/embeddings"
	"github.com/siriusfreak/hack-zurich-2023/backend/internal/templater"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func getChats(c *gin.Context) {
	chats, err := db.GetChatIDs()
	if err != nil {
		c.JSON(500, gin.H{"status": err})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "success", "chats": chats})
	}
}

func getChatById(c *gin.Context) {
	chatID := c.Param("chatID")
	chatIDParsed, err := strconv.Atoi(chatID)
	if err != nil {
		c.JSON(500, gin.H{"status": err})
		return
	}

	messages, err := db.GetChatMessages(chatIDParsed)
	if err != nil {
		c.JSON(500, gin.H{"status": err})
	}

	c.JSON(200, messages)
}

func postToExistingChat(c *gin.Context, tmpl *templater.Templater, msg db.ChatMessage, messages []db.ChatMessage) {
	allMessages := make([]chatgpt.Message, 0, len(messages)+1)
	for _, m := range messages {
		role := "assistant"
		if !m.IsBot {
			role = "user"
		}

		message := m.Message
		if m.RealMessage != "" {
			message = m.RealMessage
		}
		allMessages = append(allMessages, chatgpt.Message{
			Role:    role,
			Content: message,
		})
	}

	userMessage, err := tmpl.ProcessTemplateAllQuestionsData(msg.Message, msg.Language)
	allMessages = append(allMessages, chatgpt.Message{
		Role:    "user",
		Content: userMessage,
	})

	resp, err := chatgpt.CallAPI(chatgpt.RequestBody{
		Model:    "gpt-3.5-turbo",
		Messages: allMessages,
	})
	if err != nil {
		c.JSON(500, gin.H{"status": err})
		return
	}

	err = db.InsertChatMessage(msg.ChatID, msg.Message, "", false)
	if err != nil {
		c.JSON(500, gin.H{"status": err})
		return
	}

	err = db.InsertChatMessage(msg.ChatID, resp.Choices[0].Message.Content, "", true)
	if err != nil {
		c.JSON(500, gin.H{"status": err})
		return
	}

	c.JSON(200, gin.H{"status": "message added", "response": resp.Choices[0].Message.Content})

}

func postToNewChat(c *gin.Context, msg db.ChatMessage, tmpl *templater.Templater) {
	embed, err := embeddings.MakePredictionRequest("hackzurich23-8200",
		embeddings.PredictRequest{
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

	err = db.InsertChatMessage(msg.ChatID, msg.Message, template, false)
	if err != nil {
		c.JSON(500, gin.H{"status": err})
		return
	}
	err = db.InsertChatMessage(msg.ChatID, resp.Choices[0].Message.Content, "", true)
	if err != nil {
		c.JSON(500, gin.H{"status": err})
		return
	}

	c.JSON(200, gin.H{"status": "message added", "response": resp.Choices[0].Message.Content})
}

func processCorner(tmpl *templater.Templater, cornerName string, msg db.ChatMessage, responses chan string) {
	question, err := tmpl.GetCornerQuestion(cornerName, msg.Message)
	if err != nil {
		fmt.Printf("error getting corner question: %v\n", err)
		return
	}

	resp, err := chatgpt.CallAPI(chatgpt.RequestBody{
		Model: "gpt-3.5-turbo",
		Messages: []chatgpt.Message{
			{
				Role:    "user",
				Content: question,
			},
		},
	})
	if err != nil {
		return
	}

	if resp.Choices[0].Message.Content == "YES" {
		cornerResponse, err := tmpl.GetCornerResponse(cornerName)
		if err != nil {
			fmt.Printf("error getting corner response: %v\n", err)
			return
		}
		responses <- cornerResponse
	}
}

func processCornerCases(c *gin.Context, msg db.ChatMessage, tmpl *templater.Templater) (string, error) {
	corners := tmpl.GetCornerNames()
	responses := make(chan string, len(corners))
	wg := &sync.WaitGroup{}
	for _, cornerName := range corners {
		func(cornerName string) {
			wg.Add(1)
			processCorner(tmpl, cornerName, msg, responses)
			wg.Done()
		}(cornerName)
	}

	wg.Wait()
	close(responses)

	res := ""
	for response := range responses {
		res += response + "\n"
	}

	return res, nil
}

func postToChat(c *gin.Context, tmpl *templater.Templater) {
	var err error

	chatID := c.Param("chatID")
	var msg db.ChatMessage
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	msg.ChatID, err = strconv.Atoi(chatID)
	if err != nil {
		c.JSON(500, gin.H{"status": err})
		return
	}

	chatMessages, err := db.GetChatMessages(msg.ChatID)
	if err != nil {
		c.JSON(500, gin.H{"status": err})
		return
	}

	//resp, err := processCornerCases(c, msg, tmpl)
	//if err != nil {
	//	return
	//}
	//if resp != "" {
	//	err = db.InsertChatMessage(msg.ChatID, msg.Message, resp, false)
	//	if err != nil {
	//		c.JSON(500, gin.H{"status": err})
	//		return
	//	}
	//
	//	err = db.InsertChatMessage(msg.ChatID, msg.Message, resp, true)
	//	if err != nil {
	//		c.JSON(500, gin.H{"status": err})
	//		return
	//	}
	//	c.JSON(200, gin.H{"status": "message added", "response": resp})
	//	return
	//}

	if len(chatMessages) == 0 {
		postToNewChat(c, msg, tmpl)
	} else {
		postToExistingChat(c, tmpl, msg, chatMessages)
	}

}

func main() {
	err := db.InitDB("chat.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.CloseDB()

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

	r.GET("/chat", getChats)
	r.GET("/chat/:chatID", getChatById)

	r.POST("/chat/:chatID", func(c *gin.Context) {
		postToChat(c, tmpl)
	})

	r.Run()
}
