
# Terraform Provider for Roxy-WI

The Terraform Provider for Roxy-WI allows you to manage Roxy-WI resources such as UDP listeners and groups.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) v1.7.0+
- Go 1.26.5+ (to build the provider)

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
VERSION=1.5.5
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

## Security

Fields marked as sensitive are hidden from Terraform CLI output, but Terraform Plugin SDK v2 still stores configured secret values in state. Use an encrypted remote state backend with strict access controls. See [Security guidance](./docs/security.md) for details.


## License

MIT License. See [LICENSE](./LICENSE) for details.
