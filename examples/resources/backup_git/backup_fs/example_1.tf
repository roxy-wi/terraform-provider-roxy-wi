provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}

resource "roxywi_backup_git" "example" {
  cred_id     = 1
  description = "Daily backup of application data"
  repo        = "git@github.com:example/haproxy_configs"
  branch      = "main"
  server_id   = 29
  period      = "daily"
  service_id  = 1
}
