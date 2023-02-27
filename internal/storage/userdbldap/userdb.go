package userdbldap

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strconv"

	"github.com/go-ldap/ldap/v3"

	"s1767.xyz/idp/internal/storage/userdb"
)

func NewUserDb(ldapServer string, ldapPort int, searchBase, searchDn, searchPw string, cafile string) (userdb.UserDb, error) {
	// create the tls config
	cacert, err := os.ReadFile(cafile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cacert)

	tlsConfig := tls.Config{
		MinVersion: tls.VersionTLS13,
		RootCAs:    caCertPool,
		ServerName: ldapServer,
	}

	// create the struct
	udb := userDb{
		server:     fmt.Sprintf("%s:%d", ldapServer, ldapPort),
		tlsConfig:  &tlsConfig,
		searchBase: searchBase,
		searchDn:   searchDn,
		searchPw:   searchPw,
	}

	// do a test connection just to make sure it's all ok
	err = udb.testBind()
	if err != nil {
		return nil, err
	}

	return &udb, nil
}

type userDb struct {
	server     string
	tlsConfig  *tls.Config
	searchBase string
	searchDn   string
	searchPw   string
}

func (udb *userDb) VerifyUser(userName, userPw string) (*userdb.User, error) {
	if len(userPw) == 0 {
		return nil, fmt.Errorf("password required: %w", userdb.ErrUserNotFound)
	}
	user, err := udb.findUser(userName, userPw)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (udb *userDb) LookupUser(userName string) (*userdb.User, error) {
	user, err := udb.findUser(userName, "")
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (udb *userDb) LookupGroup(groupName string) (*userdb.Group, error) {
	group, err := udb.findGroup(groupName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w %w", groupName, userdb.ErrGroupNotFound, err)
	}
	return group, nil
}

func (udb *userDb) testBind() error {
	l, err := ldap.Dial("tcp", udb.server)
	if err != nil {
		return err
	}
	defer l.Close()
	return nil
}

func (udb *userDb) findUser(userName, userPw string) (*userdb.User, error) {
	l, err := ldap.Dial("tcp", udb.server)
	if err != nil {
		return nil, err
	}
	defer l.Close()

	// upgrade to tls
	err = l.StartTLS(udb.tlsConfig)
	if err != nil {
		return nil, err
	}

	// bind with the search user
	err = l.Bind(udb.searchDn, udb.searchPw)
	if err != nil {
		return nil, err
	}

	// search for the user
	searchRequest := ldap.NewSearchRequest(
		udb.searchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=posixAccount)(uid=%s))", userName),
		[]string{"dn", "uidNumber", "gidNumber", "cn", "sn", "givenName", "mail"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	if len(sr.Entries) != 1 {
		return nil, fmt.Errorf("%s: %w", userName, userdb.ErrUserNotFound)
	}
	userDn := sr.Entries[0].DN

	// rebind as the user if the password is given
	if len(userPw) != 0 {
		err = l.Bind(userDn, userPw)
		if err != nil {
			return nil, err
		}
	}

	// create the user object
	user := userdb.User{
		Dn:        userDn,
		Name:      userName,
		UidNumber: -1,
		GidNumber: -1,
		Password:  "redacted",
	}

	for _, attr := range sr.Entries[0].Attributes {
		switch attr.Name {
		case "uidNumber":
			uidNumber, err := strconv.Atoi(attr.Values[0])
			if err != nil {
				return nil, err
			}
			user.UidNumber = uidNumber
		case "gidNumber":
			gidNumber, err := strconv.Atoi(attr.Values[0])
			if err != nil {
				return nil, err
			}
			user.GidNumber = gidNumber
		case "cn":
			user.FullName = attr.Values[0]
		case "sn":
			user.FamilyName = attr.Values[0]
		case "givenName":
			user.GivenName = attr.Values[0]
		case "mail":
			user.Email = attr.Values[0]
		}
	}

	// search for the groups the user is part of
	searchRequest = ldap.NewSearchRequest(
		udb.searchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=posixGroup)(memberUid=%s))", userName),
		[]string{"dn", "cn"},
		nil,
	)
	sr, err = l.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var groups []string
	for _, attr := range sr.Entries[0].Attributes {
		if attr.Name == "cn" {
			for _, v := range attr.Values {
				groups = append(groups, v)
			}
			break
		}
	}
	user.Groups = groups

	return &user, nil
}

func (udb *userDb) findGroup(groupName string) (*userdb.Group, error) {
	l, err := ldap.Dial("tcp", udb.server)
	if err != nil {
		return nil, err
	}
	defer l.Close()

	// upgrade to tls
	err = l.StartTLS(udb.tlsConfig)
	if err != nil {
		return nil, err
	}

	// bind with the search user
	err = l.Bind(udb.searchDn, udb.searchPw)
	if err != nil {
		return nil, err
	}

	// search for the user
	searchRequest := ldap.NewSearchRequest(
		udb.searchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=posixGroup)(cn=%s))", groupName),
		[]string{"dn", "gidNumber"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	if len(sr.Entries) != 1 {
		return nil, fmt.Errorf("%s: %w", groupName, userdb.ErrGroupNotFound)
	}

	group := &userdb.Group{
		Dn:        sr.Entries[0].DN,
		Name:      groupName,
		GidNumber: -1,
	}

	for _, attr := range sr.Entries[0].Attributes {
		if attr.Name == "gidNumber" {
			gidNumber, err := strconv.Atoi(attr.Values[0])
			if err != nil {
				return nil, err
			}
			group.GidNumber = gidNumber
			break
		}
	}

	return group, nil
}
