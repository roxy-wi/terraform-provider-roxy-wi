
# Terraform Provider for Roxy-WI

Manage Roxy-WI servers, service configurations, HA clusters, backups and Let's Encrypt certificates with Terraform.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) v1.7.0+
- Go 1.26.6+ (to build the provider)

## Building The Provider

Clone the repository and build the provider using the Go toolchain:

```sh
git clone https://github.com/roxy-wi/terraform-provider-roxy-wi.git
cd terraform-provider-roxy-wi
go build -o terraform-provider-roxywi
```

## Installing The Provider

Move the binary into the Terraform plugins directory:

```sh
VERSION=1.6.0
PLUGIN_DIR="$HOME/.terraform.d/plugins/registry.terraform.io/Roxy-wi/roxywi/${VERSION}/linux_amd64"
mkdir -p "$PLUGIN_DIR"
mv terraform-provider-roxywi "$PLUGIN_DIR/"
```

## Using The Provider

To use the provider, include it in your Terraform configuration:

```hcl
provider "roxywi" {
  base_url = "https://demo.roxy-wi.org/"
  login    = "your-login"
  password = "your-password"
}
```

## Roxy-WI 9.1 support

Provider 1.6.0 supports the certificate lifecycle and backup scheduler available in Roxy-WI 9.1 and later.

- Certificate create, update and delete wait for Operations to finish. Failed tasks fail the apply; timeouts keep the resource ID so an ongoing task can be inspected before retrying.
- `dns_profile_id` binds an existing DNS profile from the same group. It cannot be combined with inline `api_key` or `api_token`. The profile's provider must match `type`.
- `draft = true` saves a certificate without issuing it. Change it to `false` to run preflight and issue. An issued certificate cannot be changed back to a draft.
- Certificate outputs include `status`, `pem_name`, `not_after`, `next_run_at`, `retry_at`, `last_task_id`, `last_error_code` and `legacy_pending`. Renewal remains managed by Roxy-WI.
- FS/S3 backups expose the computed `schedule` block: `timezone`, `next_run_at`, `retry_at`, `migration_required`, `last_task_id` and `last_status`. Supported periods are `hourly`, `daily`, `weekly` and `monthly`.
- Service installation and HA reconfiguration wait for task IDs returned by the API. Update Roxy-WI to include the API task receipts; installations returning only a resource ID retain the previous behavior. A Tools settings update does not reinstall a service.

```hcl
resource "roxywi_letsencrypt" "edge" {
  server_id      = 1
  domains        = ["example.com", "*.example.com"]
  type           = "cloudflare"
  dns_profile_id = 4
  draft          = true
}
```

Credentials are never refreshed from API responses. Imported resources can retain credentials stored in Roxy-WI without copying them into Terraform state. Deleting a certificate stops its managed renewal; Roxy-WI keeps deployed PEM files that services may still use.

## Security

Fields marked as sensitive are hidden from Terraform CLI output, but Terraform Plugin SDK v2 still stores configured secret values in state. Use an encrypted remote state backend with strict access controls. See [Security guidance](./docs/security.md) for details.


## License

MIT License. See [LICENSE](./LICENSE) for details.
