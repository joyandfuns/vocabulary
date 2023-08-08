// db_operations.go
package main

import (
	"database/sql"
	"fmt"
)

func tableExists(db *sql.DB, tableName string) (bool, error) {
    query := fmt.Sprintf("SELECT 1 FROM %s LIMIT 1", tableName)
    _, err := db.Exec(query)
    if err != nil {
        if err.Error() == fmt.Sprintf("Error 1146: Table 'your_dbname.%s' doesn't exist", tableName) {
            return false, nil
        }
        return false, err
    }
    return true, nil
}

func createTables(db *sql.DB) error {
	tables := map[string]string{
		"word_family": `CREATE TABLE word_family (
		    family_id INT AUTO_INCREMENT PRIMARY KEY,
		    name VARCHAR(255) NOT NULL
		)`,
		"words": `CREATE TABLE words (
		    word_id INT AUTO_INCREMENT PRIMARY KEY,
		    spelling VARCHAR(255) NOT NULL,
		    family_id INT,
		    FOREIGN KEY (family_id) REFERENCES word_family(family_id)
		)`,
		"meanings": `CREATE TABLE meanings (
		    meaning_id INT AUTO_INCREMENT PRIMARY KEY,
		    word_id INT,
		    definition TEXT NOT NULL,
		    FOREIGN KEY (word_id) REFERENCES words(word_id)
		)`,
		"examples": `CREATE TABLE examples (
		    example_id INT AUTO_INCREMENT PRIMARY KEY,
		    meaning_id INT,
		    sentence TEXT NOT NULL,
		    FOREIGN KEY (meaning_id) REFERENCES meanings(meaning_id)
		)`,
	}

	for tblName, tblSQL := range tables {
		exists, err := tableExists(db, tblName)
		if err != nil {
			return err
		}
		if !exists {
			_, err := db.Exec(tblSQL)
			if err != nil {
				return err
			}
			fmt.Printf("Table %s created successfully!\n", tblName)
		} else {
			fmt.Printf("Table %s already exists.\n", tblName)
		}
	}

	return nil
}
