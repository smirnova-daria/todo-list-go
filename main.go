package main

import (
	"net/http"
	"strconv"

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
	r.GET("/tasks", GetTasks)
	r.GET("/tasks/:id", GetTaskByID)
	r.Run()
}

func GetTasks(c *gin.Context) {
	c.JSON(http.StatusOK, tasks)
}

func GetTaskByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id parameter"})
		return
	}
	task, ok := findTaskByID(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func findTaskByID(id int) (Task, bool) {
	for _, task := range tasks {
		if task.ID == id {
			return task, true
		}
	}
	return Task{}, false
}
