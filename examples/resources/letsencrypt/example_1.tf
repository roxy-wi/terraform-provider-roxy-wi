provider "roxywi" {
  base_url = "https://..."
  login    = "testlog"
  password = "testpass"
}

resource "roxywi_letsencrypt" "example" {
  email     = "test23@gmail.com"
  domains   = ["exmaple.com", "example2.com"]
  type      = "route53"
  api_key   = "aws_access_key_id"
  api_token = "aws_secret_access_key"
  server_id = 1
}
