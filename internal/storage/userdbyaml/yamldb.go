package userdbyaml

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"

	"s1767.xyz/idp/internal/storage/userdb"
)

func NewUserDb(path string) (userdb.UserDb, error) {
	// load the config
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	decoder := yaml.NewDecoder(fh)

	var usercfg userConfig
	err = decoder.Decode(&usercfg)
	if err != nil {
		return nil, err
	}

	ydb := yamlDb{}
	ydb.users = make(map[string]userdb.User)
	ydb.groups = make(map[string]userdb.Group)
	for _, user := range usercfg.Users {
		ydb.users[user.Name] = user
	}
	for _, group := range usercfg.Groups {
		ydb.groups[group.Name] = group
	}

	return &ydb, nil
}

type userConfig struct {
	Users  []userdb.User  `yaml:"users"`
	Groups []userdb.Group `yaml:"groups"`
}

type yamlDb struct {
	users  map[string]userdb.User
	groups map[string]userdb.Group
}

func (ydb *yamlDb) VerifyUser(userName, password string) (*userdb.User, error) {
	user, ok := ydb.users[userName]
	if !ok {
		return nil, fmt.Errorf("%s: %w", userName, userdb.ErrUserNotFound)
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", userName, userdb.ErrUserNotFound)
	}
	return &user, nil
}

func (ydb *yamlDb) LookupUser(userName string) (*userdb.User, error) {
	user, ok := ydb.users[userName]
	if !ok {
		return nil, fmt.Errorf("%s: %w", userName, userdb.ErrUserNotFound)
	}
	return &user, nil
}

func (ydb *yamlDb) LookupGroup(groupName string) (*userdb.Group, error) {
	group, ok := ydb.groups[groupName]
	if !ok {
		return nil, fmt.Errorf("%s: %w", groupName, userdb.ErrGroupNotFound)
	}
	return &group, nil
}
