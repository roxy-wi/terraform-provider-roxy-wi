provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}

resource "roxywi_service_installation" "example" {
  service   = "haproxy"
  server_id = 123

  auto_start = true
  checker    = true
  metrics    = true
  syn_flood  = false
  docker     = false
}

output "service_installation_id" {
  value = roxywi_service_installation.example.id
}