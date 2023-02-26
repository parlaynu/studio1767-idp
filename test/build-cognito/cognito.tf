resource "aws_cognito_user_pool" "s1767" {
  name = var.project_code
}

resource "aws_cognito_user_pool_domain" "s1767" {
  domain       = var.project_code
  user_pool_id = aws_cognito_user_pool.s1767.id
}

resource "aws_cognito_user_pool_client" "s1767" {
  for_each = var.clients
  
  name                                 = each.key
  user_pool_id                         = aws_cognito_user_pool.s1767.id
  callback_urls                        = local.clients[each.key].redirect_urls
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows                  = ["code"]
  allowed_oauth_scopes                 = ["openid", "email", "profile"]
  supported_identity_providers         = ["COGNITO"]
  generate_secret                      = true
}



resource "aws_cognito_resource_server" "s1767" {
  for_each = var.clients
  
  name = each.key
  identifier = "https://${each.value.ip_addresses[0]}:${each.value.listen_port}"

  user_pool_id = aws_cognito_user_pool.s1767.id
}

resource "aws_cognito_user_group" "s1767" {
  for_each = var.groups
  user_pool_id = aws_cognito_user_pool.s1767.id
  name = each.key
}

resource "aws_cognito_user" "s1767" {
  for_each = var.users
  user_pool_id = aws_cognito_user_pool.s1767.id
  username     = each.key
  password     = each.value.password

  attributes = {
    email          = each.value.email
    email_verified = true
  }
}

locals {
  user_in_group = merge([ for uname, user in var.users: {
    for gname, group in var.groups: join("_", [uname,gname]) => [uname, gname] if startswith(gname, "group") || gname == uname
    }]...)
}

resource "aws_cognito_user_in_group" "s1767" {
  for_each = local.user_in_group

  user_pool_id = aws_cognito_user_pool.s1767.id
  username     = each.value[0]
  group_name   = each.value[1]
  
  depends_on = [
    aws_cognito_user.s1767,
    aws_cognito_user_group.s1767
  ]
}

