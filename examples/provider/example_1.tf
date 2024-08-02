terraform {
  required_providers {
    roxywi = {
      source = "Roxy-wi/roxywi"
    }
  }
}

provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}