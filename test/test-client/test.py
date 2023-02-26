#!/usr/bin/env python3
import os, sys
import argparse
import getpass
import requests


def authcookie(testname, sess, url, ca_crt):
    print(f"{testname}: test with cookie")
    r = sess.get(url, verify=ca_crt)
    return r
    

def authbasic(testname, sess, url, ca_crt, username, password):
    # connect to the api server
    #   will be redirected to the idp server and be presented with a form
    print(f"{testname}: starting test: {url}")
    r = sess.get(url, verify=ca_crt)

    # post back to the idp server with the passed in credentials
    payload = {"name": username, "password": password}
    print(f"{testname}: posting credentials")
    r = sess.post(r.url, params=payload, verify=ca_crt)

    return r


def test_authbasic(url, ca_crt, username, password):
    # test00: existing user, correct password
    with requests.Session() as s:
        testname = "test00"
        r = authbasic(testname, s, url, ca_crt, username, password)
        if r.status_code == 200:
            print(f"{testname}: passed: {r.status_code} {r.text.strip()}")
        else:
            print(f"{testname}: failed: {r.status_code}")

        r = authcookie(testname, s, url, ca_crt)
        if r.status_code == 200 and len(r.history) == 0:
            print(f"{testname}: passed: {r.status_code} {r.text.strip()}")
        else:
            print(f"{testname}: failed: {r.status_code} {r.history}")
    
    # test01: existing user, incorrect password
    with requests.Session() as s:
        testname = "test01"
        r = authbasic(testname, s, url, ca_crt, username, "incorrect")
        if r.status_code == 401:
            print(f"{testname}: passed: {r.status_code}")
        else:
            print(f"{testname}: failed: {r.status_code}")

    # test00: non-existing user, incorrect password
    with requests.Session() as s:
        testname = "test02"
        r = authbasic(testname, s, url, ca_crt, "nobody", "incorrect")
        if r.status_code == 401:
            print(f"{testname}: passed: {r.status_code}")
        else:
            print(f"{testname}: failed: {r.status_code}")



def authmtls(sess, url, ca_crt, user_key, user_crt):
    # connect to the api server
    #   should handle the complete flow internally
    r = sess.get(url, verify=ca_crt, cert=(user_crt, user_key))
    return r


def test_authmtls(url, ca_crt, user_keys, user_crts):
    user_ids = list(user_keys.keys())
    user_ids.sort()
    
    for idx, user_id in enumerate(user_ids):
        testname = f"test1{idx}"
        
        ukey = user_keys[user_id]
        ucrt = user_crts[user_id]
        
        print(f"{testname}: {user_id}: starting test: {url}")
        
        with requests.Session() as s:
            r = authmtls(s, url, ca_crt, ukey, ucrt)
            if user_id.endswith(".user1"):
                # this user isn't in the system... expect a 401 response
                if r.status_code == 401:
                    print(f"{testname}: {user_id}: passed: {r.status_code}")
                else:
                    print(f"{testname}: {user_id}: failed: {r.status_code}")
                continue
                
            else:
                # these users should exist so expect 200 response
                if r.status_code == 200:
                    print(f"{testname}: {user_id}: passed: {r.status_code} {r.text.strip()}")
                else:
                    print(f"{testname}: {user_id}: failed: {r.status_code}")

                r = authcookie(testname, s, url, ca_crt)
                if r.status_code == 200 and len(r.history) == 0:
                    print(f"{testname}: passed: {r.status_code} {r.text.strip()}")
                else:
                    print(f"{testname}: failed: {r.status_code} {r.history}")
            
    
    



def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('certs', help='path to certificate files', type=str, default=None)
    parser.add_argument('url', help='url to the test server', type=str, default=None)
    args = parser.parse_args()
    
    # load the certificates and keys
    ca_crt = ""
    user_keys = {}
    user_crts = {}
    
    for entry in os.scandir(args.certs):
        if entry.name == "ca.crt":
            ca_crt = entry.path
            continue
            
        if entry.name.startswith("user.") and entry.name.endswith(".key"):
            user = os.path.splitext(entry.name)[0]
            user_keys[user] = entry.path
            continue

        if entry.name.startswith("user.") and entry.name.endswith(".crt"):
            user = os.path.splitext(entry.name)[0]
            user_crts[user] = entry.path
            continue
            
    # get a username and password for basic auth testing
    username = "user2"      # input("Username: ")
    password = "password2"  # getpass.getpass("Password: ")
    
    # run an auth test
    test_authbasic(args.url, ca_crt, username, password)
    test_authmtls(args.url, ca_crt, user_keys, user_crts)


if __name__ == "__main__":
    main()

