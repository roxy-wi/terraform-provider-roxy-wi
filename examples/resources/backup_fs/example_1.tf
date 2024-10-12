provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}

resource "roxywi_backup_fs" "example" {
  cred_id     = 1
  description = "Daily backup of application data"
  rpath       = "/tmp/backup1"
  rserver     = "127.0.0.3"
  server_id   = 29
  time        = "daily"
  type        = "synchronization"
}
