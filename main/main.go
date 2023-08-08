package main

import (
	"net/http"

	"database/sql"

	"github.com/gin-gonic/gin"

	_ "github.com/go-sql-driver/mysql"

	"log"
)

type WordInfo struct {
	Spelling string    `json:"spelling"`
	Meanings []Meaning `json:"meanings"`
	Family   []string  `json:"family"`
}

type Meaning struct {
	Definition string   `json:"definition"`
	Examples   []string `json:"examples"`
}

var db *sql.DB

// use gin to listener on port 9000
func main() {
	// Database connection
	var err error
	dsn := "root:Mysql_19940620@tcp(localhost:3306)/vocabulary"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})
	router.GET("/list", list)
	router.GET("/new_tables", newTables)
	router.GET("/search/:word", searchHandler)
	router.Run(":9000")
}

// A function process /list request
func list(c *gin.Context) {
	c.String(http.StatusOK, "List")
}

func newTables(c *gin.Context) {
	if err := createTables(db); err != nil {
		c.String(http.StatusInternalServerError, "create error" + err.Error())
	} else {
		c.String(http.StatusOK, "createTables success")
	}
}

func searchHandler(c *gin.Context) {
	word := c.Param("word")

	var wordID, familyID int
	err := db.QueryRow("SELECT word_id, family_id FROM words WHERE spelling=?", word).Scan(&wordID, &familyID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Word not found"})
		return
	}

	wordInfo := WordInfo{Spelling: word}

	// Fetch meanings and examples
	rows, err := db.Query("SELECT definition FROM meanings WHERE word_id=?", wordID)
	if err != nil {
		c.String(http.StatusInternalServerError, "search error" + err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var definition string
		err = rows.Scan(&definition)
		if err != nil {
			c.String(http.StatusInternalServerError, "search error" + err.Error())
			return
		}

		// For each meaning, fetch its examples
		var examples []string
		exampleRows, err := db.Query("SELECT sentence FROM examples WHERE meaning_id=?", wordID)
		if err != nil {
			c.String(http.StatusInternalServerError, "search error" + err.Error())
			return
		}
		for exampleRows.Next() {
			var sentence string
			err = exampleRows.Scan(&sentence)
			if err != nil {
				c.String(http.StatusInternalServerError, "search error" + err.Error())
				return
			}
			examples = append(examples, sentence)
		}
		exampleRows.Close()

		meaning := Meaning{
			Definition: definition,
			Examples:   examples,
		}
		wordInfo.Meanings = append(wordInfo.Meanings, meaning)
	}

	// Fetch word family
	familyRows, err := db.Query("SELECT spelling FROM words WHERE family_id=?", familyID)
	if err != nil {
		c.String(http.StatusInternalServerError, "search error" + err.Error())
		return
	}
	defer familyRows.Close()

	for familyRows.Next() {
		var familyWord string
		err = familyRows.Scan(&familyWord)
		if err != nil {
			c.String(http.StatusInternalServerError, "search error" + err.Error())
			return
		}
		wordInfo.Family = append(wordInfo.Family, familyWord)
	}

	// Send as JSON
	c.JSON(200, wordInfo)
}