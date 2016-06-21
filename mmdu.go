package main

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"os"
	"fmt"
	"regexp"
	"strings"
	"log"
	"github.com/BurntSushi/toml"
	"reflect"
)

const selectUsersCurrentData = "select User,Host,Password,Grant_priv from mysql.user where User=? and Host=?"
const selectAllUsers = "select User, Host from mysql.user;"

func getUserFromDatabase(username, host string, db *sql.DB) (User, error) {

	var grantPriv, grantLine string
	var user User

	err := db.QueryRow(selectUsersCurrentData, username, host).Scan(&user.Username, &user.Host, &user.Password, &grantPriv)
	if err != nil {
		fmt.Println("Error querying " + selectUsersCurrentData + ": ", err.Error())
		return user, err
	} else {
		if grantPriv == "Y" {
			user.GrantOptions = true
		}
		err = db.QueryRow(fmt.Sprint("show grants for " + username + "@" + "'" + host + "'")).Scan(&grantLine)
		if err != nil {
			fmt.Println(err.Error())
			return user, err
		} else {
			re := regexp.MustCompile("GRANT (.*) ON (.*) TO.*")
			user.Privileges = strings.Split(re.ReplaceAllString(grantLine, "$1"), ",")
			user.ObjectAccess = re.ReplaceAllString(grantLine, "$2")
		}
	}
	return user, nil
}

func getAllUsersFromDB(db * sql.DB) []User {
	var users []User
	rows, err := db.Query(selectAllUsers)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var username, host string
		if err := rows.Scan(&username, &host); err != nil {
			log.Fatal(err)
		}
		user, err := getUserFromDatabase(username, host, db)
		if err == nil {
			users = append(users, user)
		}

	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return users
}

func getAllUsersFromConfig() []User {
	var users UsersConfig
	if _, err := toml.DecodeFile("./mmdu.toml", &users); err != nil {
		fmt.Println("Failed to parse config file", err.Error())
	}
	return users.User
}

func getUsersToRemove(usersFromConf, usersFromDB []User) []User {
	var usersToRemove [] User
	for _, userDB := range usersFromDB {
		var found bool
		for _, userConf := range usersFromConf {
			if reflect.DeepEqual(userConf, userDB) {
				found = true
				break
			}
		}
		if ! found {
			usersToRemove = append(usersToRemove, userDB)
		}
	}
	return usersToRemove
}

func getUsersToAdd(usersFromConf, usersFromDB []User) []User {
	var usersToAdd [] User
	for _, userConf := range usersFromConf {
		var found bool
		for _, userDB := range usersFromDB {
			if reflect.DeepEqual(userConf, userDB) {
				found = true
				break
			}
		}
		if ! found {
			usersToAdd = append(usersToAdd, userConf)
		}
	}
	return usersToAdd
}

func main() {
	db, err := sql.Open("mysql", "root@/")
	if err != nil {
		os.Exit(1)
	}
	defer db.Close()

	usersFromConf := getAllUsersFromConfig()
	usersFromDB := getAllUsersFromDB(db)

	usersToRemove := getUsersToRemove(usersFromConf, usersFromDB)
	usersToAdd := getUsersToAdd(usersFromConf, usersFromDB)

	fmt.Println(usersToAdd)
	fmt.Println(usersToRemove)
}
