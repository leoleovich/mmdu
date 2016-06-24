package main

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"crypto/sha1"
	"io"
)

type User struct {
	Username       string
	Network        string
	Password       string
	HashedPassword string
	Database       string
	Table          string
	Privileges     []string
	GrantOption    bool
}

func getUserFromDatabase(username, host string, db *sql.DB) (User, error) {

	var grantPriv, grantLine string
	var user User
	query := "SELECT User, Host, Password, Grant_priv FROM mysql.user WHERE User='" + username + "' and Host='" + host + "'"
	err := db.QueryRow(query).Scan(&user.Username, &user.Network, &user.HashedPassword, &grantPriv)
	if err != nil {
		fmt.Println("Error querying "+query+": ", err.Error())
		return user, err
	} else {
		if grantPriv == "Y" {
			user.GrantOption = true
		}
		err = db.QueryRow(fmt.Sprint("SHOW GRANTS FOR '" + username + "'@'" + host + "'")).Scan(&grantLine)
		if err != nil {
			fmt.Println(err.Error())
			return user, err
		} else {
			re := regexp.MustCompile("GRANT (.*) ON (.*)\\.(.*) TO.*")
			user.Privileges = strings.Split(re.ReplaceAllString(grantLine, "$1"), ",")
			user.Database = re.ReplaceAllString(grantLine, "$2")
			user.Table = re.ReplaceAllString(grantLine, "$3")
		}
	}

	return user, nil
}

func getAllUsersFromDB(db *sql.DB) ([]User, error) {
	var users []User
	rows, err := db.Query(selectAllUsers)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var username, host string
		if err := rows.Scan(&username, &host); err != nil {
			log.Fatal(err)
		}
		user, err := getUserFromDatabase(username, host, db)
		if err != nil {
			return users, err
		} else {
			users = append(users, user)
		}

	}
	if err := rows.Err(); err != nil {
		return users, err
	}

	return users, nil
}

func (u *User) calcUserHashPassword () {
	h := sha1.New()
	io.WriteString(h, u.Password)
	h2 := sha1.New()
	h2.Write(h.Sum(nil))

	u.HashedPassword = strings.ToUpper(strings.Replace(fmt.Sprintf("*% x", h2.Sum(nil)), " ","", -1))
}

func validateUsers(users []User) []User {
	var resultUsers []User
	for _, u := range users {
		if u.Username != "" && u.Network != "" && u.Database != "" && u.Database != "" && u.Table != "" && len(u.Privileges) > 0 {
			if u.Password != "" {
				u.calcUserHashPassword()
			}
			resultUsers = append(resultUsers, u)
		}
	}
	return resultUsers
}

func getUsersToRemove(usersFromConf, usersFromDB []User) []User {
	var usersToRemove []User
	for _, userDB := range usersFromDB {
		var found bool
		for _, userConf := range usersFromConf {
			if reflect.DeepEqual(userConf, userDB) {
				found = true
				break
			}
		}
		if !found {
			usersToRemove = append(usersToRemove, userDB)
		}
	}

	return usersToRemove
}

func getUsersToAdd(usersFromConf, usersFromDB []User) []User {
	var usersToAdd []User
	for _, userConf := range usersFromConf {
		var found bool
		for _, userDB := range usersFromDB {
			if reflect.DeepEqual(userConf, userDB) {
				found = true
				break
			}
		}
		if !found {
			usersToAdd = append(usersToAdd, userConf)
		}
	}

	return usersToAdd
}

func (u *User) dropUser(tx *sql.Tx, execute bool) bool {
	query := "DROP USER '" + u.Username + "'@'" + u.Network + "'"
	if execute {
		_, err := tx.Exec(query)
		if err != nil {
			return false
		}
	} else {
		fmt.Println(query)
	}

	return true
}

func (u *User) addUser(tx *sql.Tx, execute bool) bool {
	database := u.Database
	table := u.Table

	if strings.Contains(u.Database, "%") {
		database = "`" + u.Database + "`"
	}

	if strings.Contains(u.Table, "%") {
		table = "`" + u.Table + "`"
	}

	query := "GRANT " + strings.Join(u.Privileges, ", ") + " ON " + database + "." + table + " TO '" +
		u.Username + "'@'" + u.Network + "' IDENTIFIED BY PASSWORD '" + u.HashedPassword + "'"

	if u.GrantOption {
		query += " WITH GRANT OPTION"
	}
	if execute {
		_, err := tx.Exec(query)
		if err != nil {
			return false
		}
	} else {
		fmt.Println(query)
	}

	return true
}
