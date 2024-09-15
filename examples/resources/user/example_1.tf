provider "roxywi" {
  base_url = "https://demo.roxy-wi.org"
  login    = "testlog"
  password = "testpass"
}

resource "roxywi_user" "example" {
  email    = "test23@yandex.ru"
  enabled  = true
  password = "testpassword"
  username = "testuser2"
}
