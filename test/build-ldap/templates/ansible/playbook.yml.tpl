---
- hosts: ldap_servers
  become: yes
  gather_facts: no
  vars:
    ansible_python_interpreter: "/usr/bin/env python3"
  tasks:
  - import_role:
      name: ${ldap_server_role}
  - import_role:
      name: ${ldap_utils_role}
  - import_role:
      name: ${ldap_users_role}

