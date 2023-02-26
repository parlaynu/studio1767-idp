
## test client key/certificate

resource "tls_private_key" "client_app" {
  for_each = var.clients
  
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_cert_request" "client_app" {
  for_each = var.clients
  
  private_key_pem = tls_private_key.client_app[each.key].private_key_pem

  subject {
    common_name  = each.value.cn
    organization = var.project_name
    country = "AU"
  }

  dns_names = each.value.dns_names
  ip_addresses = each.value.ip_addresses
  uris = each.value.uris
}

resource "tls_locally_signed_cert" "client_app" {
  for_each = tls_cert_request.client_app

  ca_private_key_pem = local.test_ca_key
  ca_cert_pem        = local.test_ca_cert
  cert_request_pem   = each.value.cert_request_pem

  validity_period_hours = 240
  early_renewal_hours = 48

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "local_file" "client_app_key" {
  for_each = tls_private_key.client_app
  
  content         = each.value.private_key_pem
  filename        = "local/configs/certs/client-app.key"
  file_permission = "0600"
}

resource "local_file" "client_app_cert" {
  for_each = tls_locally_signed_cert.client_app
  
  content         = each.value.cert_pem
  filename        = "local/configs/certs/client-app.crt"
  file_permission = "0644"
}


## test app server config

resource "random_password" "client_secret" {
  for_each = var.clients
  length  = 24
  special = true
}

resource "random_id" "state_secret" {
  for_each = var.clients
  byte_length  = 32
}

resource "random_id" "cookie_hash_secret" {
  for_each = var.clients
  byte_length  = 32
}

resource "random_id" "cookie_enc_secret" {
  for_each = var.clients
  byte_length  = 16
}

resource "local_file" "app_config" {
  for_each = var.clients
  
  content = templatefile("../${each.key}/configs/config.yaml.tpl", {
    client_id       = each.key
    listener        = "https://${each.value.ip_addresses[0]}:${each.value.listen_port}"
    ca_cert_file    = "certs/ca.crt"
    https_key_file  = "certs/client-app.key"
    https_cert_file = "certs/client-app.crt"
    redirect_urls   = local.clients[each.key].redirect_urls
    client_secret   = random_password.client_secret[each.key].result
    state_secret    = random_id.state_secret[each.key].b64_std
    cookie_hash_secret = random_id.cookie_hash_secret[each.key].b64_std
    cookie_enc_secret  = random_id.cookie_enc_secret[each.key].b64_std
    idp_issuer_url    = "https://${var.service_idp.ip_addresses[0]}:${var.service_idp.backend_port}"
    idp_ca_cert_file  = "certs/ca.crt"
  })
  filename        = "local/configs/config-${each.key}.yaml"
  file_permission = "0640"
}

locals {
  clients = { for client, data in var.clients: client => {
    secret        = random_password.client_secret[client].result
    redirect_urls = flatten([for url in data.redirect_urls: 
      [
        for address in data.ip_addresses: format("https://${address}:${data.listen_port}${url}")
      ]
    ])
  }}
}
