package userdbyaml_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"

	"s1767.xyz/idp/internal/storage/userdb"
	"s1767.xyz/idp/internal/storage/userdbyaml"
)

type testUser struct {
	Name     string
	Password string
}

type testGroup struct {
	Name string
}

func TestUserDb(t *testing.T) {

	users := []testUser{
		{
			Name:     "user1",
			Password: "password1",
		},
		{
			Name:     "user2",
			Password: "password2",
		},
	}
	groups := []testGroup{
		{
			Name: "group1",
		},
		{
			Name: "group2",
		},
	}

	dbpath, err := createUserDb(users, groups)
	require.NoError(t, err)
	defer os.Remove(dbpath)

	udb, err := userdbyaml.NewUserDb(dbpath)
	require.NoError(t, err)

	// test successful attempts
	{
		for _, user := range users {
			u, err := udb.VerifyUser(user.Name, user.Password)
			require.NoError(t, err)
			require.NotNil(t, u)
			u, err = udb.LookupUser(user.Name)
			require.NoError(t, err)
			require.NotNil(t, u)
		}
		for _, group := range groups {
			g, err := udb.LookupGroup(group.Name)
			require.NoError(t, err)
			require.NotNil(t, g)
		}
	}

	// test failures
	for _, user := range users {
		u, err := udb.VerifyUser(user.Name, "incorrect-password")
		require.Error(t, err)
		require.Nil(t, u)
	}
	{
		u, err := udb.LookupUser("non-existing-user")
		require.Error(t, err)
		require.Nil(t, u)
	}
	{
		g, err := udb.LookupGroup("non-existing-group")
		require.Error(t, err)
		require.Nil(t, g)
	}
}

func createUserDb(tusers []testUser, tgroups []testGroup) (string, error) {
	// create some users
	passwords := []string{}
	users := []userdb.User{}
	for i, user := range tusers {
		crypt_pw, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
		if err != nil {
			return "", err
		}

		passwords = append(passwords, user.Password)

		users = append(users, userdb.User{
			UidNumber:  1000 + i,
			GidNumber:  1000 + i,
			Name:       user.Name,
			Password:   string(crypt_pw),
			FullName:   fmt.Sprintf("User%d Family%d", i, i),
			GivenName:  fmt.Sprintf("User%d", i),
			FamilyName: fmt.Sprintf("Family%d", i),
			Email:      fmt.Sprintf("user%d@example.com", i),
			Groups:     []string{"group0", "group1"},
		})
	}

	// create some groups
	groups := []userdb.Group{}
	for i, group := range tgroups {
		groups = append(groups, userdb.Group{
			GidNumber: 2000 + i,
			Name:      group.Name,
		})
	}

	// create the configuration
	cfg := struct {
		Users  []userdb.User
		Groups []userdb.Group
	}{
		Users:  users,
		Groups: groups,
	}

	// write to the config file
	fh, err := os.CreateTemp("", "userdb_test")
	if err != nil {
		return "", err
	}

	encoder := yaml.NewEncoder(fh)
	err = encoder.Encode(cfg)
	if err != nil {
		return "", err
	}

	name := fh.Name()
	err = fh.Close()
	if err != nil {
		return "", err
	}

	return name, nil
}
