package userdbldap

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strconv"

	"github.com/go-ldap/ldap/v3"

	"github.com/parlaynu/studio1767-oidc-idp/internal/storage/userdb"
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
	ldp := ldapDb{
		server:     fmt.Sprintf("%s:%d", ldapServer, ldapPort),
		tlsConfig:  &tlsConfig,
		searchBase: searchBase,
		searchDn:   searchDn,
		searchPw:   searchPw,
	}

	// do a test connection just to make sure it's all ok
	err = ldp.testBind()
	if err != nil {
		return nil, err
	}

	return &ldp, nil
}

type ldapDb struct {
	server     string
	tlsConfig  *tls.Config
	searchBase string
	searchDn   string
	searchPw   string
}

func (ldp *ldapDb) VerifyUser(userName, userPw string) (*userdb.User, error) {
	if len(userPw) == 0 {
		return nil, fmt.Errorf("password required: %w", userdb.ErrUserNotFound)
	}
	user, err := ldp.findUser(userName, userPw)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ldp *ldapDb) LookupUser(userName string) (*userdb.User, error) {
	user, err := ldp.findUser(userName, "")
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ldp *ldapDb) LookupGroup(groupName string) (*userdb.Group, error) {
	group, err := ldp.findGroup(groupName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w %w", groupName, userdb.ErrGroupNotFound, err)
	}
	return group, nil
}

func (ldp *ldapDb) testBind() error {
	l, err := ldap.Dial("tcp", ldp.server)
	if err != nil {
		return err
	}
	defer l.Close()
	return nil
}

func (ldp *ldapDb) findUser(userName, userPw string) (*userdb.User, error) {
	l, err := ldap.Dial("tcp", ldp.server)
	if err != nil {
		return nil, err
	}
	defer l.Close()

	// upgrade to tls
	err = l.StartTLS(ldp.tlsConfig)
	if err != nil {
		return nil, err
	}

	// bind with the search user
	err = l.Bind(ldp.searchDn, ldp.searchPw)
	if err != nil {
		return nil, err
	}

	// search for the user
	searchRequest := ldap.NewSearchRequest(
		ldp.searchBase,
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
		ldp.searchBase,
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

func (ldp *ldapDb) findGroup(groupName string) (*userdb.Group, error) {
	l, err := ldap.Dial("tcp", ldp.server)
	if err != nil {
		return nil, err
	}
	defer l.Close()

	// upgrade to tls
	err = l.StartTLS(ldp.tlsConfig)
	if err != nil {
		return nil, err
	}

	// bind with the search user
	err = l.Bind(ldp.searchDn, ldp.searchPw)
	if err != nil {
		return nil, err
	}

	// search for the user
	searchRequest := ldap.NewSearchRequest(
		ldp.searchBase,
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
