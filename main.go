package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Task struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

var tasks = []Task{
	{ID: 1, Text: "todo number 1", Done: true},
	{ID: 2, Text: "todo №2", Done: false},
}

func main() {
	r := gin.Default()
	r.GET("/tasks", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, tasks)
	})
	r.Run()
}
