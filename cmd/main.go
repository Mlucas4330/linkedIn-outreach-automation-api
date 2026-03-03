package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
)

type GenerateRequest struct {
	RawText string `json:"raw_taw"`
	Method  string `json:"text"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	llm, err := ollama.New(
		ollama.WithServerURL("http://ollama:11434"),
		ollama.WithModel("codellama"),
	)

	if err != nil {
		logger.Error("error" + err.Error())
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:            "redis:6379",
		Password:        "", // no password
		DB:              0,  // use default DB
		Protocol:        2,
		MinRetryBackoff: 10 * time.Millisecond,
		MaxRetryBackoff: 100 * time.Millisecond,
		MaxRetries:      5,
		DialTimeout:     10 * time.Second,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
	})
	defer rdb.Close()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Error("error" + err.Error())
	}

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.GET("/api/jobs/recent", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "message"})
	})

	router.GET("/api/jobs/:selectedId", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "message"})
	})

	router.POST("/api/generate", func(c *gin.Context) {
		var req GenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("error:", err.Error())
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		if req.Method == "text" {
			template := prompts.NewPromptTemplate(
				"You are a Recruiter analyst {{.cv}}. Answer this question: {{.job_description}}",
				[]string{"cv", "job_description"},
			)

			prompt, err := template.Format(map[string]any{
				"cv":              "helpful assistant",
				"job_description": req.RawText,
			})

			if err != nil {
				logger.Error("error" + err.Error())
			}

			completion, err := llms.GenerateFromSinglePrompt(
				c.Request.Context(),
				llm,
				prompt,
				llms.WithTemperature(0.8),
			)
			if err != nil {
				logger.Error("error" + err.Error())
			}

			fmt.Println(completion)

			rdb.Set(c.Request.Context(), "teste", "completition", 10*time.Minute)
		}

		c.JSON(http.StatusOK, gin.H{"message": "message"})
	})
	router.Run(":80")
}
