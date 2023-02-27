## idp server test key/certificate

resource "tls_private_key" "service_idp" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_cert_request" "service_idp" {
  private_key_pem = tls_private_key.service_idp.private_key_pem

  subject {
    common_name  = var.service_idp.cn
    organization = var.project_name
    country = "AU"
  }

  dns_names = var.service_idp.dns_names
  ip_addresses = var.service_idp.ip_addresses
  uris = var.service_idp.uris
}

resource "tls_locally_signed_cert" "service_idp" {
  ca_private_key_pem = local.test_ca_key
   ca_cert_pem        = local.test_ca_cert
  cert_request_pem   = tls_cert_request.service_idp.cert_request_pem

  validity_period_hours = 240
  early_renewal_hours = 48

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "local_file" "service_idp_key" {
  content         = tls_private_key.service_idp.private_key_pem
  filename        = "local/configs/certs/service-idp.key"
  file_permission = "0600"
}

resource "local_file" "service_idp_cert" {
  content         = tls_locally_signed_cert.service_idp.cert_pem
  filename        = "local/configs/certs/service-idp.crt"
  file_permission = "0644"
}


## idp server test config

locals {
  search_user = element(sort(tolist(keys(var.users))), 0)
}

resource "local_file" "idp_config" {
  content = templatefile("../../configs/config.yaml.tpl", {
    frontend_listen  = "https://${var.service_idp.ip_addresses[0]}:${var.service_idp.frontend_port}"
    backend_listen   = "https://${var.service_idp.ip_addresses[0]}:${var.service_idp.backend_port}"
    ca_cert_file     = "certs/ca.crt"
    https_key_file   = "certs/service-idp.key"
    https_cert_file  = "certs/service-idp.crt"
    content_dir      = join("/", [dirname(dirname(abspath(path.root))), "web"])
    clients          = local.clients
    user_db_type     = "ldap"
    user_db_file     = ""
    ldap_server      = aws_instance.s1767.public_ip
    ldap_port        = var.service_ldap.port
    ldap_search_base = local.project_domain_dn
    ldap_search_dn   = "uid=${local.search_user},ou=users,${local.project_domain_dn}"
    ldap_search_pw   = var.users[local.search_user].password
  })
  filename        = "local/configs/config-idp.yaml"
  file_permission = "0640"
}

