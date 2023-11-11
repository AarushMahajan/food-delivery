package config

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db *sql.DB
)

func Connect() {

	d, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/inventory")
	if err != nil {
		fmt.Println("error1")
		panic(err)
	}

	d.SetConnMaxIdleTime(time.Second * 5000)
	d.SetMaxOpenConns(100)
	d.SetMaxIdleConns(50)

	err = d.Ping()
	if err != nil {
		fmt.Println("error2")
		panic(err.Error())
	}

	db = d
	fmt.Println("Database connected successfully")
}

func GetDB() *sql.DB {
	return db
}
