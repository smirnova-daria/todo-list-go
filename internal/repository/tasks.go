package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type Task struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

type TaskRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) GetByID(id string) *sql.Row {
	row := r.db.QueryRow("SELECT id, text, done FROM tasks WHERE id = ?", id)
	return row
}

func (r *TaskRepository) GetAll() (*sql.Rows, error) {
	rows, err := r.db.Query("SELECT id, text, done FROM tasks")
	if err != nil {
		log.Printf("Get tasks in repository, err: %v\n", err.Error())
		return nil, err
	}
	// defer rows.Close()
	return rows, nil
}

func (r *TaskRepository) Exists(id string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tasks WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		log.Printf("Exists in repository, id=%s %v\n", id, err)
	}
	return exists, err
}

func (r *TaskRepository) Create(text string, done bool) (int64, error) {
	result, err := r.db.Exec("INSERT INTO tasks (text, done) VALUES (?, ?)", text, done)
	if err != nil {
		log.Printf("Create in repository, err: %v\n", err.Error())
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func (r *TaskRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		log.Printf("Delete from repository, err: %v\n", err.Error())
		return err
	}
	return nil
}
func (r *TaskRepository) Update(id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := "UPDATE tasks SET "
	args := []interface{}{}
	setClauses := []string{}

	allowedFields := map[string]bool{"text": true, "done": true}
	for field, value := range updates {
		if !allowedFields[field] {
			log.Printf("Update in repository, invalid field: %s", field)
			return fmt.Errorf("invalid field '%s'", field)
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", field))
		args = append(args, value)
	}

	query += strings.Join(setClauses, ", ") + " WHERE id = ?"
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		log.Printf("Update in repository, err: %v", err.Error())
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}
