provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}

resource "roxywi_backup_s3" "example" {
  access_key  = "your_access_key"
  secret_key  = "your_secret_key"
  bucket      = "your_bucket_name"
  description = "Daily backup of application data to S3"
  server_id   = 29
  s3_server   = "https://s3-server.com"
  time        = "daily"
}
