
resource "tls_private_key" "users" {
  for_each = var.users

  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_cert_request" "users" {
  for_each = var.users

  private_key_pem = tls_private_key.users[each.key].private_key_pem

  subject {
    common_name  = "${each.key}"
    organization = var.project_name
    country = "AU"
  }

  uris = ["email:${each.value.email}"]
}

resource "tls_locally_signed_cert" "users" {
  for_each = tls_cert_request.users

  ca_private_key_pem = local.test_ca_key
  ca_cert_pem        = local.test_ca_cert
  cert_request_pem   = each.value.cert_request_pem

  validity_period_hours = 240
  early_renewal_hours = 48

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "client_auth",
  ]
}

resource "local_file" "user_key" {
  for_each = tls_private_key.users

  content         = each.value.private_key_pem
  filename        = "local/configs/certs/user.${each.key}.key"
  file_permission = "0600"
}

resource "local_file" "user_cert" {
  for_each = tls_locally_signed_cert.users

  content         = each.value.cert_pem
  filename        = "local/configs/certs/user.${each.key}.crt"
  file_permission = "0644"
}

resource "null_resource" "user_pkcs12" {
  for_each = var.users

  triggers = {
    certificate = tls_locally_signed_cert.users[each.key].cert_pem
  }
  
  provisioner "local-exec" {
    command = <<-CMD
      openssl pkcs12 -export -inkey ${basename(local_file.user_key[each.key].filename)}   \
                      -in ${basename(local_file.user_cert[each.key].filename)}            \
                      -name ${var.users[each.key].given_name}_${var.users[each.key].family_name} \
                      -out user.${each.key}.${var.project_code}.pfx   \
                      -passout pass:
      CMD

    working_dir = "local/configs/certs"
  }
}

resource "null_resource" "user_der" {
  for_each = var.users

  triggers = {
    certificate = tls_locally_signed_cert.users[each.key].cert_pem
  }
  
  provisioner "local-exec" {
    command = <<-CMD
      openssl x509 -in ${basename(local_file.user_cert[each.key].filename)} \
                   -out user.${each.key}.${var.project_code}.der            \
                   -outform DER
      CMD

    working_dir = "local/configs/certs"
  }
}

locals {
  drop_user = keys(var.users)[0]
  real_users = { for name, user in var.users: name => user if name != local.drop_user}
}

resource "local_file" "user_db" {
  content = templatefile("../../configs/users.yaml.tpl", {
    users     = local.real_users
    groups    = var.groups
  })
  filename        = "local/configs/userdb.yaml"
  file_permission = "0600"
}
