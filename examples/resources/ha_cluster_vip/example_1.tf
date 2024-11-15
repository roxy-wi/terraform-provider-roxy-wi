provider "roxywi" {
  base_url = "https://..."
  login    = "your_login"
  password = "your_password"
}

resource "roxywi_ha_cluster_vip" "example" {
  virt_server = true
  use_src     = true
  vip         = "10.0.0.128"
  cluster_id  = roxywi_ha_cluster.example.id

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
}
