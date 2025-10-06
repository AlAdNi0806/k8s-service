// internal/model/user.go
package model

type User struct {
	ID       int64  `pg:"id,pk"`
	Email    string `pg:"email,unique:notnull"`
	Password string `pg:"password,notnull"`
}
