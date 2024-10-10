provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}

resource "roxywi_service_installation" "example" {
  service   = "haproxy"
  server_id = 123

  auto_start = 1
  checker    = 1
  metrics    = 1
  syn_flood  = 0
  docker     = 0
}

output "service_installation_id" {
  value = roxywi_service_installation.example.id
}