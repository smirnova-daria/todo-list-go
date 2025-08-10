package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smirnova-daria/todo-list-go/internal/repository"
)

type TaskHandler struct {
	repo *repository.TaskRepository
}

type TaskUpdateRequest struct {
	Text *string `json:"text"` // Указатель, чтобы отличать "не передан" от "пусто"
	Done *bool   `json:"done"`
}

func New(repo *repository.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	exists, err := h.repo.Exists(id)
	if err != nil {
		log.Printf("handler DeleteTask, err: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error, try later"})
		return
	}
	if !exists {
		log.Printf("handler DeleteTask, task does not exists\n")
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		log.Printf("handler DeleteTask, err: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error, try later"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "successfully deleted"})
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	rows, err := h.repo.GetAll()
	if err != nil {
		log.Printf("handler GetTasks, err: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something goes wrong, try later"})
		return
	}
	defer rows.Close()
	var tasks []repository.Task
	for rows.Next() {
		var task repository.Task
		err := rows.Scan(&task.ID, &task.Text, &task.Done)
		if err != nil {
			log.Printf("handler GetTasks, err: %v\n", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "something goes wrong, try later"})
			return
		}
		tasks = append(tasks, task)
	}
	if tasks == nil {
		tasks = []repository.Task{}
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var task repository.Task
	if err := c.BindJSON(&task); err != nil {
		log.Printf("handler CreateTask, empty request\n")
		c.JSON(http.StatusBadRequest, gin.H{"message": "you provided empty task"})
		return
	}

	if task.Text == "" || strings.TrimSpace(task.Text) == "" {
		log.Printf("handler CreateTask, invalid task text")
		c.JSON(http.StatusBadRequest, gin.H{"message": "you provided empty task text"})
		return
	}

	id, err := h.repo.Create(task.Text, task.Done)
	if err != nil {
		log.Printf("handler CreateTask, err: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something goes wrong, try later"})
		return
	}

	task.ID = int(id)

	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	exists, err := h.repo.Exists(id)
	if err != nil {
		log.Printf("handler UpdateTask, err: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error, try later"})
		return
	}
	if !exists {
		log.Printf("handler UpdateTask, task does not exists\n")
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	var req TaskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("handler UpdateTask, err: %v\n", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "you provided invalid data"})
		return
	}

	updates := make(map[string]interface{})

	if req.Text != nil {
		if strings.TrimSpace(*req.Text) == "" {
			log.Printf("handler UpdateTask, empty text field\n")
			c.JSON(http.StatusBadRequest, gin.H{"error": "field text can't be empty"})
			return
		}
		updates["text"] = *req.Text
	}

	if req.Done != nil {
		updates["done"] = *req.Done
	}

	err = h.repo.Update(id, updates)
	if err != nil {
		log.Printf("handler UpdateTask, err: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error, try later"})
		return
	}

	var task repository.Task
	row := h.repo.GetByID(id)
	row.Scan(&task.ID, &task.Text, &task.Done)
	c.JSON(http.StatusOK, task)
}
