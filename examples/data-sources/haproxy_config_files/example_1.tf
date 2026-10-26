data "roxywi_haproxy_config_files" "example" {
  server_id = 1
}

output "haproxy_config_files" {
  value = data.roxywi_haproxy_config_files.example.files
}

output "haproxy_multiple_files_enabled" {
  value = data.roxywi_haproxy_config_files.example.multiple_files
}
