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
