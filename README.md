
# Terraform Provider for Roxy-WI

The Terraform Provider for Roxy-WI allows you to manage Roxy-WI resources such as UDP listeners and groups.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) v1.7.0+
- Go 1.22.5+ (to build the provider)

## Building The Provider

Clone the repository and build the provider using the Go toolchain:

```sh
git clone <your-repository-url>
cd <your-repository-directory>
go build -o terraform-provider-roxywi
```

## Installing The Provider

Move the binary into the Terraform plugins directory:

```sh
mkdir -p ~/.terraform.d/plugins/roxywi.com/roxywi/1.0.0/linux_amd64
mv terraform-provider-roxywi ~/.terraform.d/plugins/roxywi.com/roxywi/1.0.0/linux_amd64
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


## License

MIT License. See [LICENSE](./LICENSE) for details.
