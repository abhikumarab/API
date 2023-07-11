package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Task represents a task entity in the database.
type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Status      string `json:"status"`
}

// Database connection
var db *sql.DB

func main() {
	// Connect to the PostgreSQL database
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=postgres password=abhikumar10 dbname=task_manager sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize Gin router
	router := gin.Default()

	// Define API endpoints
	router.POST("/tasks", createTask)
	router.GET("/tasks/:id", getTask)
	router.PUT("/tasks/:id", updateTask)
	router.DELETE("/tasks/:id", deleteTask)
	router.GET("/tasks", listTasks)

	// Start the server
	router.Run(":8080")
}

// Handler for creating a new task
func createTask(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	err := db.QueryRow(`
		INSERT INTO tasks (title, description, due_date, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`, task.Title, task.Description, task.DueDate, task.Status).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	task.ID = id

	c.JSON(http.StatusCreated, task)
}

// Handler for retrieving a task by ID
func getTask(c *gin.Context) {
	id := c.Param("id")

	var task Task
	err := db.QueryRow(`
		SELECT id, title, description, due_date, status
		FROM tasks WHERE id = $1;
	`, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// Handler for updating a task by ID
func updateTask(c *gin.Context) {
	id := c.Param("id")

	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec(`
		UPDATE tasks
		SET title = $1, description = $2, due_date = $3, status = $4
		WHERE id = $5;
	`, task.Title, task.Description, task.DueDate, task.Status, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// Handler for deleting a task by ID
func deleteTask(c *gin.Context) {
	id := c.Param("id")

	result, err := db.Exec("DELETE FROM tasks WHERE id = $1;", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// Handler for listing all tasks
func listTasks(c *gin.Context) {
	rows, err := db.Query(`
		SELECT id, title, description, due_date, status
		FROM tasks;
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	tasks := make([]Task, 0)
	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, tasks)
}
