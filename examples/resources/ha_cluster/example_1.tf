provider "roxywi" {
  base_url = "https://..."
  login    = "your_login"
  password = "your_password"
}

resource "roxywi_ha_cluster" "example" {
  description = "Example HA"
  virt_server = true
  use_src     = false
  name        = "example listener"
  syn_flood   = false
  vip         = "10.0.0.127"

  servers {
    id     = 1
    eth    = "eth0"
    master = true
  }

  servers {
    id     = 29
    eth    = "eth0"
    master = false
  }

  services {
    name    = "haproxy"
    docker  = false
    enabled = true
  }

  services {
    name    = "nginx"
    docker  = false
    enabled = false
  }

  services {
    name    = "apache"
    docker  = false
    enabled = false
  }
}
