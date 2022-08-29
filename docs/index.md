---
page_title: "Provider: RDS Configuration"
description: |-
  The RDS Configuration provider manages RDS settings in an AWS RDS MySQL
  instance or Aurora MySQL cluster.
---

# RDS Configuration Provider

The RDS Configuration provider manages RDS settings in an AWS RDS MySQL instance
or Aurora MySQL cluster. These configuration settings cannot be controlled by
the AWS API or console UI. Instead, they are maintained in the MySQL database
itself and are managed with the `mysql.rds_show_configuration()` and
`mysql.rds_set_configuration()` stored procedures.

Thus, this provider requires credentials to log into the database in order to
make these stored procedure calls. If you are already managing the
`aws_db_instance` or `aws_rds_cluster` in Terraform, you can use the same values
you provide for `username`/`password` or `master_username`/`master_password` in
these resources, respectively. You can also use the `endpoint` attribute from
these resources (or data sources) for this provider.

## Single Configuration Resource per Provider

You can create multiple instances of the RDS Configuration provider to manage
several individual RDS instances or Aurora clusters; however, you can (should)
only create one `rds-configuration_configuration` resource for each
provider. While you technically _can_ create multiple configuration resources,
if any of them managed the same settings, they will collide, flipping the values
with each `apply`.

As a reminder of this, all `rds-configuration_configuration` resources have an
`id` of `singleton`.

## Example Usage

### Provider

```terraform
provider "rds-configuration" {
  endpoint = "aurora-foo.cluster-ygEZvJQyhzKJ.us-west-2.rds.amazonaws.com"
  username = "root"
  password = "foobarbazblah"
}

# configure multiple providers using aliases
provider "rds-configuration" {
  alias    = "bar"
  endpoint = "aurora-bar.cluster-ygEZvJQyhzKJ.us-west-2.rds.amazonaws.com"
  username = "root"
  password = "barbazblahfoo"
}

# configure provider from environment variable
provider "rds-configuration" {
  alias = "baz"
  # connection information pulled from environment variables:
  # `MYSQL_ENDPOINT`, `MYSQL_USERNAME`, `MYSQL_PASSWORD`, etc.
}

# configure provider from Terraform resources or data sources
provider "rds-configuration" {
  alias    = "blah"
  endpoint = aws_rds_cluster.blah.endpoint        # or `data.aws_rds_cluster.blah.endpoint` from data source
  username = aws_rds_cluster.blah.master_username # or `data.aws_rds_cluster.blah.master_username`
  password = random_password.blah.result
}
```

### Data Source

The RDS Configuration data source simply pulls the results of the
`mysql.rds_show_configuration()` stored procedure into Terraform. This will
provide a list of all of the supported settings as well as their
descriptions. The `setting` attribute is a list of maps with the keys `name`,
`value`, and `description`.

```terraform
data "rds-configuration_configuration" "foo" {}

data "rds-configuration_configuration" "bar" {
  provider = rds-configuration.bar
}
```

### Resource

The RDS Configuration resource manages one or more settings — there are only a
few settings supported in RDS/Aurora — via `setting` blocks with attributes
`name` and `value`.

```terraform
resource "rds-configuration_configuration" "foo" {
  setting {
    name  = "binlog retention hours"
    value = 7 * 24
  }
}

resource "rds-configuration_configuration" "bar" {
  provider = rds-configuration.bar

  setting {
    name  = "binlog retention hours"
    value = 3 * 24
  }
}
```

### Importing

An RDS Configuration resouce can be imported by providing any value for `id`,
but as stated above the provider will always use `singleton` for the `id`.

```shell
terraform import 'rds-configuration_configuration.foo' 'singleton'
```
