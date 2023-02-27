#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "$${BASH_SOURCE[0]}" )" && pwd )"
cd $${SCRIPT_DIR}

PWFILE=/etc/ldap/secrets/admin-pw.txt

if [ -r $${PWFILE} ]; then
  ldapwhoami -H ldap://${ldap_server}:${ldap_port} -ZZ -x -D cn=admin,${domain_dn} -y $${PWFILE}
else
  ldapwhoami -H ldap://${ldap_server}:${ldap_port} -ZZ -x -D cn=admin,${domain_dn} -W
fi

