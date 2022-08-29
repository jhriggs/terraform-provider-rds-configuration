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
