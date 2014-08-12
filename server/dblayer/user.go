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

func (us UserService) Update(user *model.User) error {
	_, err := gDb.Update(user)
	return err
}

func (us UserService) ByUsername(username string) (*model.User, bool) {
	user := &model.User{}
	err := gDb.SelectOne(user, UserByUsernameSql, username)
	if err != nil {
		return nil, false
	}
	return user, true
}
