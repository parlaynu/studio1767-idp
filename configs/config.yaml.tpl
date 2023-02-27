
listeners:
  frontend: ${frontend_listen}
  backend: ${backend_listen}

https:
  ca_cert_file: ${ca_cert_file}
  key_file: ${https_key_file}
  cert_file: ${https_cert_file}

content_dir: ${content_dir}

clients:
%{ for id, client in clients ~}
- id: ${id}
  secret: "${client.secret}"
  redirect_urls:
%{ for url in client.redirect_urls ~}
  - ${url}
%{ endfor ~}
%{ endfor ~}

user_db:
  type: ${user_db_type}
  path: ${user_db_file}
  ldap_server: ${ldap_server}
  ldap_port: ${ldap_port}
  search_base: ${ldap_search_base}
  search_dn: ${ldap_search_dn}
  search_pw: ${ldap_search_pw}

