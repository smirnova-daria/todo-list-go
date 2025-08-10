package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/smirnova-daria/todo-list-go/internal/handlers"
	"github.com/smirnova-daria/todo-list-go/internal/repository"
)

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

	taskRepo := repository.New(db)
	taskHandler := handlers.New(taskRepo)

	r := gin.Default()
	r.GET("/tasks", taskHandler.GetTasks)
	r.POST("/tasks", taskHandler.CreateTask)
	r.DELETE("/tasks/:id", taskHandler.DeleteTask)
	r.PATCH("/tasks/:id", taskHandler.UpdateTask)

	r.Run()
}
