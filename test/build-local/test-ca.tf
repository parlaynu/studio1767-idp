
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
    common_name  = "Test Root CA"
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

resource "local_file" "ca_bundle" {
  for_each = var.clients
  
  content         = local.test_ca_cert
  filename        = "local/configs/certs/ca.crt"
  file_permission = "0644"
}
