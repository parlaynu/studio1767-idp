
## Root CA key and certificate

locals {
  test_ca_key = tls_private_key.test_ca.private_key_pem
  test_ca_cert = tls_self_signed_cert.test_ca.cert_pem
}

resource "tls_private_key" "test_ca" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_self_signed_cert" "test_ca" {
  private_key_pem = tls_private_key.test_ca.private_key_pem

  subject {
    common_name  = "Test Cognito CA"
    organization = var.project_name
    country = "AU"
  }
  
  validity_period_hours = 480
  early_renewal_hours = 48

  allowed_uses = [
    "cert_signing",
    "crl_signing"
  ]
  is_ca_certificate = true
}

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

resource "local_file" "client_app_ca_bundle" {
  for_each = var.clients
  
  content         = local.test_ca_cert
  filename        = "local/configs/certs/ca.crt"
  file_permission = "0644"
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
