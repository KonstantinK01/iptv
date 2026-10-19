package dto

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	Active       bool
}
