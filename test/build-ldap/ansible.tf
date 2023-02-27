## local variables

locals {
  ldap_server_role = "ldap-server"
  ldap_utils_role = "ldap-utils"
  ldap_users_role = "ldap-users"
  ldap_server = aws_instance.s1767.tags["Name"]
}

## render the run script

resource "local_file" "run_playbook" {
  content = templatefile("templates/ansible/run-ansible.sh.tpl", {
      inventory_file = "inventory.ini"
    })
  filename = "local/ansible/run-ansible.sh"
  file_permission = "0755"
}


## render the playbook

resource "local_file" "playbook" {
  content = templatefile("templates/ansible/playbook.yml.tpl", {
      ldap_server_role = local.ldap_server_role,
      ldap_utils_role = local.ldap_utils_role,
      ldap_users_role = local.ldap_users_role,
    })
  filename = "local/ansible/playbook.yml"
  file_permission = "0640"
}


## render the inventory file

resource "local_file" "inventory" {
  content = templatefile("templates/ansible/inventory.ini.tpl", {
    ldap_server = local.ldap_server,
  })
  filename = "local/ansible/inventory.ini"
  file_permission = "0640"
}


## hostvars for ldap server

resource "local_file" "ldap_host_vars" {
  content = templatefile("templates/ansible/host_vars/ldap-server.yml.tpl", {
    server_name = local.ldap_server
    organization = lower(var.project_name)
    domain_dns = var.project_domain
    domain_dn = local.project_domain_dn
    domain_dc = split(".", var.project_domain)[0]
    users  = var.users
    groups = var.groups
  })
  filename        = "local/ansible/host_vars/${local.ldap_server}.yml"
  file_permission = "0640"
}


## render the roles

resource "random_password" "ldap" {
  length           = 16
  special          = true
}

resource "template_dir" "ldap_server" {
  source_dir      = "templates/ansible-roles/${local.ldap_server_role}"
  destination_dir = "local/ansible/roles/${local.ldap_server_role}"

  vars = {
    ldap_port   = var.service_ldap.port
    ca_cert     = local.test_ca_cert
    server_cert = tls_locally_signed_cert.s1767_ldap.cert_pem
    server_key  = tls_private_key.s1767_ldap.private_key_pem
    admin_password = random_password.ldap.result
  }
}

resource "template_dir" "ldap_utils" {
  source_dir      = "templates/ansible-roles/${local.ldap_utils_role}"
  destination_dir = "local/ansible/roles/${local.ldap_utils_role}"

  vars = {
    ldap_server = "127.0.0.1"
    ldap_port   = var.service_ldap.port
    domain_dns  = var.project_domain
    domain_dn   = local.project_domain_dn
    ca_cert_bundle = local.test_ca_cert
    admin_password = random_password.ldap.result
  }
}

resource "template_dir" "ldap_users" {
  source_dir      = "templates/ansible-roles/${local.ldap_users_role}"
  destination_dir = "local/ansible/roles/${local.ldap_users_role}"

  vars = {
    ldap_port   = var.service_ldap.port
    domain_dn   = local.project_domain_dn
    admin_password = random_password.ldap.result
  }
}

