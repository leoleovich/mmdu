package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

type General struct {
	AutoExecute bool
}

type Config struct {
	General  General
	Access   Access
	Database []Database
	User     []User
}

const selectAllUsers = "SELECT User, Host FROM mysql.user"
const showAllDatabases = "SHOW DATABASES"

func main() {
	var execute bool
	var configFile string

	flag.BoolVar(&execute, "e", false, "Execute. If specified - changes will be applied")
	flag.StringVar(&configFile, "c", "/etc/mmdu/mmdu.toml", "Path to config file.")
	flag.Parse()

	var conf Config
	if _, err := toml.DecodeFile(configFile, &conf); err != nil {
		fmt.Println("Failed to parse config file", err.Error())
		os.Exit(1)
	}

	// Execute statements ether on -e argument or autoExecute from config file
	if conf.General.AutoExecute == true {
		execute = true
	}

	defaultDatabases := []Database{{"information_schema"}, {"mysql"},
		{"performance_schema"}}
	validatedUsers, err := validateUsers(conf.User)
	if err != nil {
		fmt.Println("Error during validation of user list:", err.Error())
		os.Exit(1)
	}

	db := conf.Access.connectAndCheck()
	defer db.Close()

	usersFromDB, err := getAllUsersFromDB(db)
	if err != nil {
		fmt.Println("Failed to get users from DB", err.Error())
		os.Exit(2)
	}

	databasesFromDB, err := getDatabasesFromDB(db)
	if err != nil {
		fmt.Println("Failed to get databases from DB", err.Error())
		os.Exit(2)
	}

	// Merge from 3 sources: default, from DbConfig and from UserConfig
	databasesFromConf := removeDuplicateDatabases(
		append(append(defaultDatabases, conf.Database...), getDatabasesFromUsers(validatedUsers)...))

	usersToRemove := getUsersToRemove(validatedUsers, usersFromDB)
	usersToAdd := getUsersToAdd(validatedUsers, usersFromDB)

	databasesToRemove := getDatabasesToRemove(databasesFromConf, databasesFromDB)
	databasesToAdd := getDatabasesToAdd(databasesFromConf, databasesFromDB)

	if len(usersToRemove) == 0 && len(usersToAdd) == 0 && len(databasesToRemove) == 0 && len(databasesToAdd) == 0 {
		fmt.Println("Nothing to do")
	} else {
		tx, err := db.Begin()
		if err != nil {
			fmt.Println("Failed to start transaction", err.Error())
			os.Exit(2)
		}

		for _, database := range databasesToRemove {
			err := database.dropDatabase(tx, execute)
			if err != nil {
				fmt.Println(database.Name, ":", err)
				tx.Rollback()
				return
			}
		}

		for _, database := range databasesToAdd {
			err := database.addDatabase(tx, execute)
			if err != nil {
				fmt.Println(database.Name, ":", err)
				tx.Rollback()
				return
			}
		}

		for _, user := range usersToRemove {
			err := user.dropUser(tx, execute)
			if err != nil {
				fmt.Println(user.Username, ":", err)
				tx.Rollback()
				return
			}
		}

		for _, user := range usersToAdd {
			err := user.addUser(tx, execute)
			if err != nil {
				fmt.Println(user.Username, ":", err)
				tx.Rollback()
				return
			}
		}
		tx.Commit()
	}
}
