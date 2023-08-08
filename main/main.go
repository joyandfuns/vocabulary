package main

import (
	"net/http"

	"database/sql"

	"github.com/gin-gonic/gin"

	_ "github.com/go-sql-driver/mysql"
)

// use gin to listener on port 9000
func main() {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})
	router.GET("/list", list)
	router.GET("/new_tables", newTables)
	router.Run(":9000")
}

// A function process /list request
func list(c *gin.Context) {
	c.String(http.StatusOK, "List")
}

func newTables(c *gin.Context) {
	// Set your database connection string here
	// Format: username:password@tcp(host:port)/dbname
	dsn := "root:Mysql_19940620@tcp(localhost:3306)/vocabulary"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		c.String(http.StatusInternalServerError, "open error" + err.Error())
	}
	defer db.Close()

	if err := createTables(db); err != nil {
		c.String(http.StatusInternalServerError, "create error" + err.Error())
	} else {
		c.String(http.StatusOK, "createTables success")
	}
}