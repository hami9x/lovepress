package model

type UserRole uint32

const (
	RoleAnonymous UserRole = 1 << iota
	RoleUser
	RoleModerator
	RoleAdmin
)

type Post struct {
	Id      int64
	Title   string
	Content string
}

type UsernamePassword struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	UsernamePassword
	Id    int64
	Role  UserRole
	Token string `json:"token"`
}
