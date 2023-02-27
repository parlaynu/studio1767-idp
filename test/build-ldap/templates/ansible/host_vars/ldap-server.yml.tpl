---
server_name: ${server_name}

organization: ${organization}
domain_dns: ${domain_dns}
domain_dn: ${domain_dn}
domain_dc: ${domain_dc}


ldap_users:
%{ for name, user in users ~}
- name: ${name}
  uid: ${user.uid}
  gid: ${user.gid}
  email: ${user.email}
  full_name: ${user.full_name}
  given_name: ${user.given_name}
  family_name: ${user.family_name}
  password: ${user.password}
  groups:
%{ for group in user.groups ~}
  - ${group}
%{ endfor ~}
%{ endfor ~}

ldap_groups:
%{ for name, group in groups ~}
- name: ${name}
  gid: ${group.gid}
%{ endfor ~}


