package dblayer

import "github.com/phaikawl/lovepress/model"

const (
	UserByUsernameSql = "SELECT * FROM users WHERE username=$1"
)

var (
	User UserService = UserService{}
)

type UserService struct{}

func (us UserService) Create(user *model.User) error {
	return gDb.Insert(user)
}

func (us UserService) ByUsername(username string) *model.User {
	user := &model.User{}
	err := gDb.SelectOne(user, UserByUsernameSql, username)
	if err != nil {
		return nil
	}
	return user
}
