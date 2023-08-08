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
	router.POST("/add", addWordHandler)
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

type WordInput struct {
	Spelling  string   `json:"spelling"`
	Family    string   `json:"family"`
	Meanings  []string `json:"meanings"`
	Examples  [][]string `json:"examples"`  // 注意，这是一个二维数组，对应于每个意义的例子
}

func addWordHandler(c *gin.Context) {
	var input WordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. 插入或查找 word_family
	var familyID int64
	err := db.QueryRow("SELECT family_id FROM word_family WHERE name=?", input.Family).Scan(&familyID)
	if err != nil {
		result, err := db.Exec("INSERT INTO word_family (name) VALUES (?)", input.Family)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert word family"})
			return
		}
		// 获取新插入的 family_id
		familyID, err = result.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve family id"})
			return
		}
	}

	// 2. 使用此家族ID插入 words 表
	result, err := db.Exec("INSERT INTO words (spelling, family_id) VALUES (?, ?)", input.Spelling, familyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert word"})
		return
	}
	wordID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve word id"})
		return
	}

	// 3. 使用单词ID插入 meanings 表
	for index, definition := range input.Meanings {
		result, err := db.Exec("INSERT INTO meanings (word_id, definition) VALUES (?, ?)", wordID, definition)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert meaning"})
			return
		}
		meaningID, err := result.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve meaning id"})
			return
		}

		// 4. 使用意义ID插入 examples 表
		for _, sentence := range input.Examples[index] {
			_, err = db.Exec("INSERT INTO examples (meaning_id, sentence) VALUES (?, ?)", meaningID, sentence)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert example"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Word added successfully"})
}
