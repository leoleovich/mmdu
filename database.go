package main

import (
	"github.com/BurntSushi/toml"
	"fmt"
	"database/sql"
	"log"
	"strings"
	"regexp"
)

type Database struct {
	Name string
}

type DatabaseConfig struct {
	Database []Database
}

func getDatabasesFromConfig() []Database {
	var databases DatabaseConfig
	if _, err := toml.DecodeFile("./mmdu.toml", &databases); err != nil {
		fmt.Println("Failed to parse config file", err.Error())
	}

	return databases.Database
}

func getDatabasesFromUsers(users []User) []Database {
	var databases []Database

	for _, user := range users {
		if user.Database == "*" {
			continue
		}
		databases = append(databases, Database{user.Database})
	}

	return databases
}

func getDatabasesFromDB(db *sql.DB) []Database  {
	var databases []Database
	rows, err := db.Query(showAllDatabases)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var database string
		if err := rows.Scan(&database); err != nil {
			log.Fatal(err)
		}
		databases = append(databases, Database{database})

	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return databases
}

func removeDuplicatesDatabases(dbs []Database) []Database {
	result := []Database{}
	seen := map[Database]bool{}
	for _, db := range dbs {
		if _, ok := seen[db]; !ok {
			result = append(result, db)
			seen[db] = true
		}
	}
	return result
}

func getDatabasesToRemove(databasesFromConf, databasesFromDB []Database) []Database {
	var databasesToRemove []Database
	for _, dbDB := range databasesFromDB {
		var found bool
		for _, dbConf := range databasesFromConf {
			if dbConf.Name == dbDB.Name || strings.Contains(dbConf.Name, "%") {
				found = true
				break
			}

		}
		if !found {
			databasesToRemove = append(databasesToRemove, dbDB)
		}
	}
	return databasesToRemove
}

func getDatabasesToAdd(databasesFromConf, databasesFromDB []Database) []Database {
	var databasesToAdd []Database
	for _, dbConf := range databasesFromConf {
		var found bool
		for _, dbDB := range databasesFromDB {
			if dbConf.Name == dbDB.Name {
				found = true
				break
			} else if strings.Contains(dbConf.Name, "%") {
				re := regexp.MustCompile(strings.Replace(dbConf.Name, "%", ".*", -1))
				if re.MatchString(dbDB.Name) {
					found = true
					break
				}
			}
		}
		if !found {
			databasesToAdd = append(databasesToAdd, dbConf)
		}
	}
	return databasesToAdd
}


func (d *Database) dropDatabase(db *sql.DB) bool {
	_, err := db.Exec("drop database " + d.Name)
	if err != nil {
		return false
	}
	return true
}

func (d *Database) addDatabase(db *sql.DB) bool {
	_, err := db.Exec("create database " + d.Name)
	if err != nil {
		return false
	}
	return true
}