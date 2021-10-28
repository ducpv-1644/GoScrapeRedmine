package config

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"strings"
)

func DBConnect() *gorm.DB {
	dsn := []string{
		"host=" + os.Getenv("POSTGRES_HOST"),
		"port=" + os.Getenv("POSTGRES_PORT"),
		"user=" + os.Getenv("POSTGRES_USER"),
		"dbname=" + os.Getenv("POSTGRES_DB"),
		"password=" + os.Getenv("POSTGRES_PASSWORD"),
		"sslmode=" + os.Getenv("POSTGRES_SSLMODE"),
	}
	db, err := gorm.Open(postgres.Open(strings.Join(dsn, " ")), &gorm.Config{})

	if err != nil {
		fmt.Println("Db connect failed!")
	}
	return db
}
