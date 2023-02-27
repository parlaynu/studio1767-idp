## management ip

data "external" "my_public_ip" {
  program = ["scripts/my-public-ip.sh"]
}

locals {
  management_ip = "${data.external.my_public_ip.result["my_public_ip"]}/32"
}

## site vpc

resource "aws_vpc" "s1767" {
  cidr_block = var.aws_cidr_block
  enable_dns_support = true
  enable_dns_hostnames = false
  tags = {
    Name = var.project_code
  }
}

## setup routing

resource "aws_default_route_table" "s1767" {
  default_route_table_id = aws_vpc.s1767.default_route_table_id
  tags = {
    Name = var.project_code
  }
}

resource "aws_route" "gateway_default" {
  route_table_id         = aws_default_route_table.s1767.id
  gateway_id             = aws_internet_gateway.s1767.id
  destination_cidr_block = "0.0.0.0/0"
}

## setup security group

resource "aws_default_security_group" "s1767" {
  vpc_id = aws_vpc.s1767.id
  tags = {
    Name = var.project_code
  }
}

resource "aws_security_group_rule" "all_out" {
  security_group_id = aws_default_security_group.s1767.id
  type        = "egress"
  protocol    = -1
  from_port   = 0
  to_port     = 0
  cidr_blocks = [ "0.0.0.0/0" ]
}

resource "aws_security_group_rule" "ssh_in" {
  security_group_id = aws_default_security_group.s1767.id
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 22
  to_port     = 22
  cidr_blocks = [ local.management_ip ]
}

resource "aws_security_group_rule" "ldap_in" {
  security_group_id = aws_default_security_group.s1767.id
  type        = "ingress"
  protocol    = "tcp"
  from_port   = var.service_ldap.port
  to_port     = var.service_ldap.port
  cidr_blocks = [ local.management_ip ]
}

## create subnet

resource "aws_subnet" "s1767" {
  vpc_id     = aws_vpc.s1767.id
  cidr_block = var.aws_cidr_block
  availability_zone = local.aws_availability_zone

  tags = {
    Name = var.project_code
  }
}

## the internet gateway

resource "aws_internet_gateway" "s1767" {
  vpc_id = aws_vpc.s1767.id
  tags = {
    Name = var.project_code
  }
}

## ssh key

resource "tls_private_key" "s1767" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "aws_key_pair" "s1767" {
  key_name   = var.project_code
  public_key = tls_private_key.s1767.public_key_openssh
}

resource "local_file" "s1767" {
  content         = tls_private_key.s1767.private_key_openssh
  filename        = "local/pki/${var.project_code}"
  file_permission = "0600"
}

resource "local_file" "ssh_config" {
  content = templatefile("templates/ssh.cfg.tpl", {
    ssh_key_file = "local/pki/${var.project_code}"
    name      = local.ldap_server
    public_ip = aws_instance.s1767.public_ip
  })
  filename        = "local/ssh.cfg"
  file_permission = "0640"
}


## the ldap server

data "aws_ami" "s1767" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "s1767" {
  instance_type = var.aws_instance_type
  ami           = data.aws_ami.s1767.id

  disable_api_termination     = false
  associate_public_ip_address = true
  source_dest_check           = true

  subnet_id                   = aws_subnet.s1767.id
  key_name                    = aws_key_pair.s1767.key_name
  vpc_security_group_ids      = [aws_default_security_group.s1767.id]

  tags = {
    Name = var.project_code
  }
  
  user_data = <<-EOF
  #!/usr/bin/env bash
  hostnamectl set-hostname ${var.project_code}
  EOF
}

## ldap server certificates

resource "tls_private_key" "s1767_ldap" {
  algorithm = "ED25519"
}

resource "tls_cert_request" "s1767_ldap" {
  private_key_pem = tls_private_key.s1767_ldap.private_key_pem

  subject {
    common_name  = "${var.project_name} LDAP Server"
    organization = var.project_name
    country = "AU"
  }

  dns_names = [ "${var.project_code}.${var.project_domain}" ]
  ip_addresses = [ "127.0.0.1", aws_instance.s1767.public_ip ]
  uris = []
}

resource "tls_locally_signed_cert" "s1767_ldap" {
  ca_private_key_pem = local.test_ca_key
  ca_cert_pem        = local.test_ca_cert
  cert_request_pem   = tls_cert_request.s1767_ldap.cert_request_pem

  validity_period_hours = 240
  early_renewal_hours = 48

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "local_file" "s1767_ldap_key" {
  content         = tls_private_key.s1767_ldap.private_key_pem
  filename        = "local/configs/certs/service-ldap.key"
  file_permission = "0600"
}

resource "local_file" "s1767_ldap_cert" {
  content         = tls_locally_signed_cert.s1767_ldap.cert_pem
  filename        = "local/configs/certs/service-ldap.crt"
  file_permission = "0644"
}
