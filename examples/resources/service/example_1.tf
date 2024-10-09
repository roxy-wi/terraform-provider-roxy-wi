provider "roxywi" {
  api_url = var.api_url
  token   = var.token
}

resource "roxywi_service_installation" "example" {
  service   = "example_service"
  server_id = 123

  auto_start = 1
  checker    = 1
  metrics    = 1
  syn_flood  = 0

  servers = [
    {
      id   = 1
      name = "server1"
    },
    {
      id   = 2
      name = "server2"
    }
  ]

  services = {
    haproxy = {
      docker  = 1
      enabled = 1
    }
  }
}

output "service_installation_id" {
  value = roxywi_service_installation.example.id
}