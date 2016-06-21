package main

import "database/sql"

type User struct {
	Username     string
	Host         string
	Password     string
	ObjectAccess string
	Privileges   []string
	GrantOptions bool
}

type UsersConfig struct {
	User []User
}

func (u *User) removeUser(db * sql.DB) bool  {
	_, err := db.Exec("drop user " + u.Username + "@" + u.Password)
	if err != nil {
		return false
	}
	return true
}