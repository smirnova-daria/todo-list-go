package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
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
	r.GET("/tasks", func(ctx *gin.Context) {
		rows, err := db.Query("SELECT id, text, done FROM tasks")
		if err != nil {
			log.Printf("get tasks, err: %v", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "something goes wrong, try later"})
			return
		}
		defer rows.Close()
		var tasks []Task
		for rows.Next() {
			var task Task
			err := rows.Scan(&task.ID, &task.Text, &task.Done)
			if err != nil {
				log.Printf("get tasks, err: %v", err.Error())
				ctx.JSON(http.StatusInternalServerError, gin.H{"message": "something goes wrong, try later"})
				return
			}
			tasks = append(tasks, task)
		}
		if tasks == nil {
			tasks = []Task{}
		}
		ctx.JSON(http.StatusOK, tasks)

	})
	r.Run()
}
