## project definitions

variable "project_name" {
  default = "Studio1767"
}

variable "project_code" {
  default = "s1767"
}

variable "users" {
  type = map(object({
    uid = number
    gid = number
    groups = list(string)
    email = string
    full_name = string
    given_name = string
    family_name = string
    password = string
  }))
  default = {
    user1 = {
      uid = 1001
      gid = 1001
      groups = ["group1", "group2"]
      email = "user1@example.xyz" 
      full_name = "one user"
      given_name = "one"
      family_name = "user"
      password = "password1"
    }
    user2 = {
      uid = 1002
      gid = 1002
      groups = ["group2"]
      email = "user2@example.xyz" 
      full_name = "two user"
      given_name = "two"
      family_name = "user"
      password = "password2"
    }
    user3 = {
      uid = 1003
      gid = 1003
      groups = ["group3"]
      email = "user3@example.xyz" 
      full_name = "three user"
      given_name = "three"
      family_name = "user"
      password = "password3"
    }
  }
}

variable "groups" {
  type = map(object({
    gid = number
  }))
  default = {
    user1 = {
      gid = 1001
    }
    user2 = {
      gid = 1002
    }
    group1 = {
      gid = 2001
    }
    group2 = {
      gid = 2002
    }
  }
}

variable "clients" {
  type = map(object({
    cn = string
    dns_names = list(string)
    ip_addresses = list(string)
    uris = list(string)
    listen_port = number
    redirect_urls = list(string)
  }))
  default = {
    test-server = {
      cn = "app.example.xyz"
      dns_names = []
      ip_addresses = ["127.0.0.1"]
      uris = []
      listen_port = 8000
      redirect_urls = ["/auth/callback"]
    }
  }
}

variable "service_idp" {
  type = object({
    cn = string
    dns_names = list(string)
    ip_addresses = list(string)
    uris = list(string)
    frontend_port = number
    backend_port = number
  })
  default = {
    cn = "idp.example.xyz"
    dns_names = ["idp-00.example.xyz", "idp-01.example.xyz"]
    ip_addresses = ["127.0.0.1"]
    uris = []      
    frontend_port = 9000
    backend_port = 9001
  }
}
