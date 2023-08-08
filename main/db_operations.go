// db_operations.go
package main

import (
	"database/sql"
	"fmt"
)

func createTables(db *sql.DB) error {
	// 注意：这里我们使用了一个数组而不是映射来确保正确的顺序
	tables := []struct {
		name string
		sql  string
	}{
		{
			"word_family",
			`CREATE TABLE IF NOT EXISTS word_family (
			    family_id INT AUTO_INCREMENT PRIMARY KEY,
			    name VARCHAR(255) NOT NULL
			)`,
		},
		{
			"words",
			`CREATE TABLE IF NOT EXISTS words (
			    word_id INT AUTO_INCREMENT PRIMARY KEY,
			    spelling VARCHAR(255) NOT NULL,
			    family_id INT,
			    FOREIGN KEY (family_id) REFERENCES word_family(family_id)
			)`,
		},
		{
			"meanings",
			`CREATE TABLE IF NOT EXISTS meanings (
			    meaning_id INT AUTO_INCREMENT PRIMARY KEY,
			    word_id INT,
			    definition TEXT NOT NULL,
			    FOREIGN KEY (word_id) REFERENCES words(word_id)
			)`,
		},
		{
			"examples",
			`CREATE TABLE IF NOT EXISTS examples (
			    example_id INT AUTO_INCREMENT PRIMARY KEY,
			    meaning_id INT,
			    sentence TEXT NOT NULL,
			    FOREIGN KEY (meaning_id) REFERENCES meanings(meaning_id)
			)`,
		},
	}

	for _, tbl := range tables {
		_, err := db.Exec(tbl.sql)
		if err != nil {
			return err
		}
		fmt.Printf("Table %s ensured to exist.\n", tbl.name)
	}

	return nil
}
