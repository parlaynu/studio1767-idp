package userdb

type UserDb interface {
	Verify(username, password string) bool
	LookupUser(username string) (*User, error)
	LookupGroup(groupname string) (*Group, error)
}

type User struct {
	UserId     int      `yaml:"uid"`
	GroupId    int      `yaml:"gid"`
	UserName   string   `yaml:"name"`
	Password   string   `yaml:"password"`
	FullName   string   `yaml:"full_name"`
	GivenName  string   `yaml:"given_name"`
	FamilyName string   `yaml:"family_name"`
	Email      string   `yaml:"email"`
	Groups     []string `yaml:"groups"`
}

type Group struct {
	GroupId int    `yaml:"gid"`
	Name    string `yaml:"name"`
}
