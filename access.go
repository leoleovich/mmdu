package main

import (
	"database/sql"
	"fmt"
	"os"
)

type Access struct {
	Username     string
	Password     string
	InitPassword string
	Host         string
	Port         int
	Socket       string
}

func (a *Access) getConnectionString(initPass bool) string {
	if a.Username == "" {
		a.Username = "root"
	}
	if a.Host == "" {
		a.Host = "localhost"
	}
	if a.Port == 0 {
		a.Port = 3306
	}
	protocol := fmt.Sprintf("tcp(%s:%d)", a.Host, a.Port)
	if a.Socket != "" {
		protocol = fmt.Sprintf("unix(%s)", a.Socket)
	}

	if initPass {
		if a.InitPassword == "" {
			return fmt.Sprintf("%s@%s/", a.Username, protocol)
		} else {
			return fmt.Sprintf("%s:%s@%s/", a.Username, a.InitPassword, protocol)
		}
	} else {
		if a.Password == "" {
			return fmt.Sprintf("%s@%s/", a.Username, protocol)
		} else {
			return fmt.Sprintf("%s:%s@%s/", a.Username, a.Password, protocol)
		}
	}

}

func (a *Access) connectAndCheck() *sql.DB {
	db, err := sql.Open("mysql", a.getConnectionString(true))
	if err != nil {
		fmt.Println("Unable to connect to mysql", err.Error())
		os.Exit(2)
	}

	// Check if we can access database with initPass
	err = db.Ping()
	if err != nil {
		db, err = sql.Open("mysql", a.getConnectionString(false))
		if err != nil {
			fmt.Println("Unable to connect to mysql", err.Error())
		}

		// Check if we can access database with regular pass
		err = db.Ping()
		if err != nil {
			fmt.Println("Can not connect to mysql with both passwords", err.Error())
			os.Exit(2)
		}
	}

	return db
}
