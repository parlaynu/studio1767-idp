users:
%{ for name, user in users ~}
- name: ${name}
  password: ${passwords[name]}
  uid: ${user.uid}
  gid: ${user.gid}
  email: ${user.email}
  full_name: ${user.full_name}
  given_name: ${user.given_name}
  family_name: ${user.family_name}
  groups:
%{ for group in user.groups ~}
  - ${group}
%{ endfor ~}
%{ endfor ~}


groups:
%{ for name, group in groups ~}
- name: ${name}
  gid: ${group.gid}
%{ endfor ~}
