resource "roxywi_haproxy_section_frontend" "example" {
  name           = "example-frontend"
  mode           = "http"
  ssl_offloading = true
  server_id      = 1
  waf            = false
  headers {
    path   = "http-response"
    method = "add-header"
    name   = "test-header"
    value  = "test"
  }
  binds {
    ip   = "0.0.0.0"
    port = 8088
  }
  acls {
    acl_if         = 1
    acl_value      = "example.com"
    acl_then       = 5
    acl_then_value = "test_backend"
  }
  acls {
    acl_if    = 2
    acl_value = "example2.com"
    acl_then  = 4
  }
}
