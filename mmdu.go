package main

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"os"
	"flag"
	"fmt"
)

const selectUsersCurrentData = "SELECT User,Host,Password,Grant_priv FROM mysql.user WHERE User=? and Host=?"
const selectAllUsers = "SELECT User, Host FROM mysql.user"
const showAllDatabases = "SHOW DATABASES"



func main() {
	var execute bool
	flag.BoolVar(&execute, "e", false, "Execute. If specified - changes will be applied")
	flag.Parse()

	db, err := sql.Open("mysql", "root@tcp(localhost:3306)/")
	if err != nil {
		os.Exit(1)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		fmt.Println("Failed to start transaction", err.Error())
		os.Exit(1)
	}


	usersFromDB, err := getAllUsersFromDB(db)
	if err != nil {
		fmt.Println("Failed during execution " + selectAllUsers, err.Error())
		os.Exit(2)
	}

	databasesFromDB, err := getDatabasesFromDB(db)
	if err != nil {
		fmt.Println("Failed during execution " + showAllDatabases, err.Error())
		os.Exit(2)
	}

	usersFromConf, err := getAllUsersFromConfig()
	if err != nil {
		fmt.Println("Failed to parse config file", err.Error())
		os.Exit(2)
	}
	databasesFromConfigDB, err := getDatabasesFromConfig()
	if err != nil {
		fmt.Println("Failed to parse config file", err.Error())
		os.Exit(2)
	}
	databasesFromConf := removeDuplicateDatabases(append(getDatabasesFromUsers(usersFromConf), databasesFromConfigDB...))

	usersToRemove := getUsersToRemove(usersFromConf, usersFromDB)
	usersToAdd := getUsersToAdd(usersFromConf, usersFromDB)

	databasesToRemove := getDatabasesToRemove(databasesFromConf, databasesFromDB)
	databasesToAdd := getDatabasesToAdd(databasesFromConf, databasesFromDB)

	for _, user := range usersToRemove {
		if ! user.dropUser(tx, execute) {
			tx.Rollback()
		}
	}

	for _, user := range usersToAdd {
		if ! user.addUser(tx, execute) {
			tx.Rollback()
		}
	}

	for _, database := range databasesToRemove {
		if ! database.dropDatabase(tx, execute) {
			tx.Rollback()
		}
	}

	for _, database := range databasesToAdd {
		if ! database.addDatabase(tx, execute) {
			tx.Rollback()
		}
	}

	tx.Commit()
}
