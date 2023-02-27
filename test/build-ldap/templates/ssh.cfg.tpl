
# gateways access
Host ldap ${name}
  Hostname ${public_ip}

Host *
  User ubuntu
  IdentityFile ${ssh_key_file}
  IdentitiesOnly yes

