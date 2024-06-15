package types

import (
	"time"
)

type User struct {
	Id        string `gorm:"primaryKey"`
	Email     string `gorm:"unique"`
	FirstName string
	LastName  string
	RoleID    *uint `gorm:"index"`
	RoleName  *string
	Role      *Role     `gorm:"foreignKey:RoleID"`
	CreatedAt time.Time `gorm:"default:now()"`
	UpdatedAt time.Time `gorm:"default:now()"`
}

type UserList []User

type CreateUserRequestBody struct {
	FirstName string
	LastName  string
	Email     string
	RoleName  string
}

type GetUserRequestBody struct {
	Email string
}

type DeleteUserRequestBody struct {
	Id    string
	Email string
}

type UpdateUserRequestBody struct {
	Id           string
	Email        string
	NewFirstName string
	NewLastName  string
	NewRoleName  string
}
