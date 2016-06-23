package main

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"os"
	"fmt"
)

const selectUsersCurrentData = "SELECT User,Host,Password,Grant_priv FROM mysql.user WHERE User=? and Host=?"
const selectAllUsers = "SELECT User, Host FROM mysql.user"
const showAllDatabases = "SHOW DATABASES"



func main() {
	db, err := sql.Open("mysql", "test:yeiDiepu1shieJee@tcp(10.0.66.31:3306)/")
	if err != nil {
		os.Exit(1)
	}
	defer db.Close()

	usersFromConf := getAllUsersFromConfig()
	usersFromDB := getAllUsersFromDB(db)

	databasesFromConf := removeDuplicatesDatabases(append(getDatabasesFromUsers(usersFromConf), getDatabasesFromConfig()...))
	databasesFromDB := getDatabasesFromDB(db)

	usersToRemove := getUsersToRemove(usersFromConf, usersFromDB)
	usersToAdd := getUsersToAdd(usersFromConf, usersFromDB)

	databasesToRemove := getDatabasesToRemove(databasesFromConf, databasesFromDB)
	databasesToAdd := getDatabasesToAdd(databasesFromConf, databasesFromDB)

	fmt.Println(usersToRemove)
	fmt.Println(usersToAdd)

	fmt.Println(databasesToRemove)
	fmt.Println(databasesToAdd)
}
