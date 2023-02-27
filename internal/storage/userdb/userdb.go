package userdb

import (
	"errors"
)

type UserDb interface {
	VerifyUser(userName, password string) (*User, error)
	LookupUser(userName string) (*User, error)
	LookupGroup(groupName string) (*Group, error)
}

type User struct {
	Dn         string
	Name       string   `yaml:"name"`
	UidNumber  int      `yaml:"uid"`
	GidNumber  int      `yaml:"gid"`
	Password   string   `yaml:"password"`
	FullName   string   `yaml:"full_name"`
	GivenName  string   `yaml:"given_name"`
	FamilyName string   `yaml:"family_name"`
	Email      string   `yaml:"email"`
	Groups     []string `yaml:"groups"`
}

type Group struct {
	Dn        string
	Name      string `yaml:"name"`
	GidNumber int    `yaml:"gid"`
}

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrGroupNotFound = errors.New("group not found")
)
