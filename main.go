package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

type TaskUpdateRequest struct {
	Text *string `json:"text"` // Указатель, чтобы отличать "не передан" от "пусто"
	Done *bool   `json:"done"`
}

func main() {
	db, err := sql.Open("sqlite3", "tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (  
			id INTEGER PRIMARY KEY AUTOINCREMENT,  
			text TEXT NOT NULL,  
			done BOOLEAN DEFAULT FALSE  
		) 
	`)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.GET("/tasks", func(ctx *gin.Context) { getTasks(ctx, db) })
	r.POST("/tasks", func(ctx *gin.Context) { createTask(ctx, db) })
	r.DELETE("/tasks/:id", func(ctx *gin.Context) { deleteTask(ctx, db) })
	r.PATCH("/tasks/:id", func(ctx *gin.Context) { editTask(ctx, db) })

	r.Run()
}

func getTasks(ctx *gin.Context, db *sql.DB) {
	rows, err := db.Query("SELECT id, text, done FROM tasks")
	if err != nil {
		log.Printf("get tasks, err: %v\n", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "something goes wrong, try later"})
		return
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Text, &task.Done)
		if err != nil {
			log.Printf("get tasks, err: %v\n", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "something goes wrong, try later"})
			return
		}
		tasks = append(tasks, task)
	}
	if tasks == nil {
		tasks = []Task{}
	}
	ctx.JSON(http.StatusOK, tasks)
}

func createTask(ctx *gin.Context, db *sql.DB) {
	var task Task
	if err := ctx.BindJSON(&task); err != nil {
		log.Printf("POST /tasks empty request")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "you provided empty task"})
		return
	}

	if task.Text == "" || strings.TrimSpace(task.Text) == "" {
		log.Printf("POST /tasks invalid task text")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "you provided empty task text"})
		return
	}

	result, err := db.Exec("INSERT INTO tasks (text, done) VALUES (?, ?)", task.Text, task.Done)
	if err != nil {
		log.Printf("POST /tasks err: %v\n", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "something goes wrong, try later"})
		return
	}
	id, _ := result.LastInsertId()
	task.ID = int(id)

	ctx.JSON(http.StatusCreated, task)
}

func deleteTask(ctx *gin.Context, db *sql.DB) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Printf("DELETE /tasks/:id err: %v\n", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "you provided invalid id"})
		return
	}
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM tasks WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		log.Printf("DELETE /tasks/:id err: %v\n", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "something goes wrong, try later"})
		return
	}
	if !exists {
		log.Printf("DELETE /tasks/:id not found task")
		ctx.JSON(http.StatusNotFound, gin.H{"message": "task with that id is not found"})
		return
	}

	_, err = db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		log.Printf("DELETE /tasks/:id err: %v\n", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "something goes wrong, try later"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "successfully deleted"})
}

func editTask(ctx *gin.Context, db *sql.DB) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Printf("PATCH /tasks/:id err: %v\n", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "you provided invalid id"})
		return
	}
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM tasks WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		log.Printf("PATCH /tasks/:id err: %v\n", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "something goes wrong, try later"})
		return
	}
	if !exists {
		log.Printf("PATCH /tasks/:id not found task")
		ctx.JSON(http.StatusNotFound, gin.H{"message": "task with that id is not found\n"})
		return
	}

	var req TaskUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "you provided invalid data"})
		return
	}

	query := "UPDATE tasks SET "
	args := []interface{}{}
	updates := []string{}

	if req.Text != nil {
		if strings.TrimSpace(*req.Text) == "" {
			log.Printf("PATCH /tasks/:id empty text field\n")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "field text can't be empty"})
			return
		}
		updates = append(updates, "text = ?")
		args = append(args, *req.Text)
	}

	if req.Done != nil {
		updates = append(updates, "done = ?")
		args = append(args, *req.Done)
	}

	if len(updates) == 0 {
		log.Printf("PATCH /tasks/:id empty fields\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "there is no fields to update"})
		return
	}

	query += strings.Join(updates, ", ") + " WHERE id = ?"
	args = append(args, id)

	_, err = db.Exec(query, args...)
	if err != nil {
		log.Printf("PATCH /tasks/:id err: %v\n", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "something goes wrong, try later"})
		return
	}

	var task Task
	row := db.QueryRow("SELECT id, text, done FROM tasks WHERE id = ?", id)
	row.Scan(&task.ID, &task.Text, &task.Done)
	ctx.JSON(http.StatusOK, task)
}
