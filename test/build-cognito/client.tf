
## test app server config

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
    client_id       = aws_cognito_user_pool_client.s1767[each.key].id
    listener        = "https://${each.value.ip_addresses[0]}:${each.value.listen_port}"
    ca_cert_file    = "certs/ca.crt"
    https_key_file  = "certs/client-app.key"
    https_cert_file = "certs/client-app.crt"
    redirect_urls   = local.clients[each.key].redirect_urls
    client_secret   = aws_cognito_user_pool_client.s1767[each.key].client_secret
    state_secret    = random_id.state_secret[each.key].b64_std
    cookie_hash_secret = random_id.cookie_hash_secret[each.key].b64_std
    cookie_enc_secret  = random_id.cookie_enc_secret[each.key].b64_std
    idp_issuer_url     = "https://${aws_cognito_user_pool.s1767.endpoint}"
    idp_ca_cert_file   = ""
  })
  filename        = "local/configs/config-${each.key}.yaml"
  file_permission = "0640"
}


locals {
  clients = { for client, data in var.clients: client => {
    redirect_urls = flatten([for url in data.redirect_urls:
      [
        for address in data.ip_addresses: format("https://${address}:${data.listen_port}${url}")
      ]
    ])
  }}
}
