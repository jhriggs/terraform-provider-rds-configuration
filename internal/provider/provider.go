package provider

import (
	"context"
	"database/sql"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	authNative    = "native"
	authCleartext = "cleartext"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			DataSourcesMap: map[string]*schema.Resource{
				"rds-configuration_configuration": dataSourceRdsConfiguration(),
			},

			ResourcesMap: map[string]*schema.Resource{
				"rds-configuration_configuration": resourceRdsConfiguration(),
			},

			Schema: map[string]*schema.Schema{
				"endpoint": &schema.Schema{
					Description: joinStrings(
						"The endpoint (IP address or hostname) of the RDS instance to ",
						"which the provider will connect. If not provided, the value of ",
						"the `MYSQL_ENDPOINT` environment variable will be used.",
					),
					Type:             schema.TypeString,
					Optional:         true,
					DefaultFunc:      schema.EnvDefaultFunc("MYSQL_ENDPOINT", nil),
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
				},
				"port": &schema.Schema{
					Description: joinStrings(
						"The TCP port of the RDS instance to which the provider will ",
						"connect. If not provided, the value of the `MYSQL_PORT`",
						"environment variable or the default of `3306` will be used.",
					),
					Type:             schema.TypeInt,
					Optional:         true,
					DefaultFunc:      schema.EnvDefaultFunc("MYSQL_PORT", nil),
					ValidateDiagFunc: validation.ToDiagFunc(validation.IsPortNumber),
				},
				"username": &schema.Schema{
					Description: joinStrings(
						"The username of the MySQL user with which to authenticate. If ",
						"not provided, the value of the `MYSQL_USERNAME` environment ",
						"variable will be used.",
					),
					Type:             schema.TypeString,
					Optional:         true,
					DefaultFunc:      schema.EnvDefaultFunc("MYSQL_USERNAME", nil),
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
				},
				"password": &schema.Schema{
					Description: joinStrings(
						"The password of the MySQL user with which to authenticate. If ",
						"not provided, the value of the `MYSQL_PASSWORD` environment ",
						"variable will be used.",
					),
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("MYSQL_PASSWORD", nil),
				},
				"tls": {
					Description: joinStrings(
						"The TLS setting with which to connect. If not provided, the ",
						"value of the `MYSQL_TLS_CONFIG` environment variable will be ",
						"used. Valid values are: `true`, `false`, `skip-verify`.",
					),
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MYSQL_TLS_CONFIG", nil),
					ValidateDiagFunc: validation.ToDiagFunc(
						validation.StringInSlice(
							[]string{
								"true",
								"false",
								"skip-verify",
							},
							false,
						),
					),
				},
				"authentication_type": {
					Description: joinStrings(
						"The authentication options with which to log in. If not ",
						"provided, the value of the `MYSQL_AUTHENTICATION_TYPE` ",
						"environment variable will be used. Valid values are: ",
						"`"+authNative+"`, `"+authCleartext+"`. The default is ",
						"`"+authNative+"`.",
					),
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MYSQL_AUTHENTICATION_TYPE", authNative),
					ValidateDiagFunc: validation.ToDiagFunc(
						validation.StringInSlice(
							[]string{
								authNative,
								authCleartext,
							},
							false,
						),
					),
				},
				"connect_timeout": &schema.Schema{
					Description: joinStrings(
						"The timeout (in seconds) for establishing a connection. The ",
						"default is `30`.",
					),
					Type:             schema.TypeInt,
					Optional:         true,
					Default:          30,
					ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
				},
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type mysqlClient struct {
	config         *mysql.Config
	connection     func(context.Context, *mysqlClient) (*sql.DB, error)
	connectTimeout time.Duration
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		endpoint := d.Get("endpoint").(string)
		user := d.Get("username").(string)
		auth := d.Get("authentication_type").(string)

		if endpoint == "" {
			return nil, diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Missing endpoint configuration",
					Detail:   "MySQL endpoint must be specified either in provider configuration or MYSQL_ENDPOINT environment variable.",
				},
			}
		}

		if user == "" {
			return nil, diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Missing username configuration",
					Detail:   "MySQL username must be specified either in provider configuration or MYSQL_USERNAME environment variable.",
				},
			}
		}

		if port := d.Get("port"); (port != nil) && (port.(int) > 0) {
			endpoint = net.JoinHostPort(endpoint, strconv.Itoa(port.(int)))
		}

		client := &mysqlClient{
			config: &mysql.Config{
				User:                    user,
				Passwd:                  d.Get("password").(string),
				Net:                     "tcp",
				Addr:                    endpoint,
				TLSConfig:               d.Get("tls").(string),
				AllowCleartextPasswords: (auth == authCleartext),
				AllowNativePasswords:    (auth == authNative),
			},
			connection:     getConnectionFunc(),
			connectTimeout: (time.Duration(d.Get("connect_timeout").(int)) * time.Second),
		}

		return client, nil
	}
}

func getConnectionFunc() func(context.Context, *mysqlClient) (*sql.DB, error) {
	var db *sql.DB

	return func(ctx context.Context, client *mysqlClient) (*sql.DB, error) {
		if db != nil {
			return db, nil
		}

		retryErr := resource.RetryContext(
			ctx,
			client.connectTimeout,
			func() *resource.RetryError {
				var err error

				db, err = sql.Open("mysql", client.config.FormatDSN())
				if err != nil {
					return resource.RetryableError(err)
				}

				err = db.Ping()
				if err != nil {
					return resource.RetryableError(err)
				}

				return nil
			},
		)
		if retryErr != nil {
			return nil, retryErr
		}

		return db, nil
	}
}

func readConfiguration(ctx context.Context, client *mysqlClient, withDescriptions bool) (map[string]map[string]interface{}, error) {
	db, err := client.connection(ctx, client)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("CALL mysql.rds_show_configuration")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var (
		name        string
		value       sql.NullInt64
		description string
	)

	result := map[string]map[string]interface{}{}

	for rows.Next() {
		err := rows.Scan(&name, &value, &description)
		if err != nil {
			return nil, err
		}

		result[name] = map[string]interface{}{
			"name":  name,
			"value": map[bool]interface{}{true: value.Int64, false: nil}[value.Valid],
		}

		if withDescriptions {
			result[name]["description"] = description
		}
	}

	return result, nil
}

func joinStrings(s ...string) string {
	return strings.Join(s, "")
}

func joinStringsSep(sep string, s ...string) string {
	return strings.Join(s, sep)
}
