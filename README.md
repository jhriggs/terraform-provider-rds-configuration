# Terraform RDS Configuration Provider

  The RDS Configuration provider manages RDS settings in an AWS RDS MySQL
  instance or Aurora MySQL cluster.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
- [Go](https://golang.org/doc/install) >= 1.18
- One or more AWS [RDS or Aurora](https://console.aws.amazon.com/rds/home) MySQL
  instances

## Building The Provider

1. Clone the repository.
2. Enter the repository directory.
3. Build the provider using the `make install` command:
```sh
$ make install
```

## Using the Provider

See `docs/index.md`.

## Developing the Provider

If you wish to work on the provider, you'll first need
[Go](http://www.golang.org) installed on your machine (see
[Requirements](#requirements) above).

To compile the provider, run `make install`. This will build the provider and
put the provider binary in the local terraform plugin directory
(e.g. `~/.terraform.d/plugins/`.

To generate or update documentation, run `make generate`. (It will also be
generated as part of `make install`.

## To Do

1. Write acceptance tests.
