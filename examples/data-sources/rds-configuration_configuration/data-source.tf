data "rds-configuration_configuration" "foo" {}

data "rds-configuration_configuration" "bar" {
  provider = rds-configuration.bar
}
