
# Terraform Provider for Roxy-WI

The Terraform Provider for Roxy-WI allows you to manage Roxy-WI resources such as UDP listeners and groups.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12+
- Go 1.12+ (to build the provider)

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
  base_url = "https://demo.roxy-wi.org/api"
  login    = "your-login"
  password = "your-password"
}

resource "roxywi_udp_listener" "example" {
  cluster_id   = 1
  name         = "example_listener"
  port         = 9997
  vip          = "192.168.1.1"
  lb_algo      = "Round robin"
  config       = [
    {
      backend_ip = "192.168.1.100"
      port       = 443
      weight     = 50
    },
    {
      backend_ip = "192.168.1.101"
      port       = 443
      weight     = 50
    }
  ]
}

data "roxywi_udp_listener" "example" {
  id = roxywi_udp_listener.example.id
}
```

## Resources

### `roxywi_udp_listener`

A resource for managing UDP listeners.

### `roxywi_group`

A resource for managing groups.

#### Arguments

- `name` (Required, String) - Name of the group.
- `description` (Optional, String) - Description of the group.


## License

MIT License. See [LICENSE](./LICENSE) for details.

