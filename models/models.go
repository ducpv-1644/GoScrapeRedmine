package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string `json:"name"`
	Email    string `gorm:"unique" json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type Authentication struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Token struct {
	Role        string `json:"role"`
	Email       string `json:"email"`
	TokenString string `json:"token"`
}

type Project struct {
	gorm.Model
	Name   string `json:"name"`
	Prefix string `json:"prefix`
}
type Member struct {
	gorm.Model
	MemberId    string `json:"menberid"`
	MemberName  string `json:"menbername"`
	MemberEmail string `json:"menberemail"`
}
type Activity struct {
	gorm.Model
	MemberId    string `json:"menberid"`
	MemberName  string `json:"menbername"`
	Project     string `json:"project"`
	Time        string `json:"time"`
	Date        string `json:"date"`
	Issues      string `json:"issues"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

func DBMigrate(db *gorm.DB) {
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Project{})
	db.AutoMigrate(&Member{})
	db.AutoMigrate(&Activity{})
}
