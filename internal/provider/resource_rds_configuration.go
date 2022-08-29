package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"golang.org/x/exp/maps"
)

func resourceRdsConfiguration() *schema.Resource {
	return &schema.Resource{
		Description: "RDS MySQL configuration settings.",

		ReadContext:   resourceRdsConfigurationRead,
		CreateContext: resourceRdsConfigurationCreateOrUpdate,
		UpdateContext: resourceRdsConfigurationCreateOrUpdate,
		DeleteContext: resourceRdsConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRdsConfigurationImport,
		},

		CustomizeDiff: resourceRdsConfigurationCustomizeDiff,

		Schema: map[string]*schema.Schema{
			"setting": {
				Description: "An individual RDS MySQL configuration setting.",
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: joinStrings(
								"The setting name. Only settings supported by RDS/Aurora will ",
								"be allowed. Others will cause an error during `plan` or ",
								"`apply`.",
							),
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Description: "The value for the setting.",
							Type:        schema.TypeInt,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func resourceRdsConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config, err := readConfiguration(ctx, meta.(*mysqlClient), false)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("setting", maps.Values(config))

	return nil
}

func resourceRdsConfigurationCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	db, err := meta.(*mysqlClient).connection(ctx, meta.(*mysqlClient))
	if err != nil {
		return diag.FromErr(err)
	}

	for _, s := range d.Get("setting").(*schema.Set).List() {
		s := s.(map[string]interface{})

		_, err := db.Exec("CALL mysql.rds_set_configuration(?, ?)", s["name"], s["value"])
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId("singleton")

	return resourceRdsConfigurationRead(ctx, d, meta)
}

func resourceRdsConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId("")

	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "RDS configuration is not actually deleted or modified in RDS",
			Detail: "Deleting an RDS configuration from Terraform will remove the " +
				"configuration from Terraforms state, but does not delete, change, " +
				"or revert anything in the RDS instance. Any previously configured " +
				"settings will persist with their most recent values.",
		},
	}
}

func resourceRdsConfigurationImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	diags := resourceRdsConfigurationRead(ctx, d, meta)
	if (diags != nil) && diags.HasError() {
		for _, d := range diags {
			if d.Severity == diag.Error {
				errMsg := d.Summary

				if d.Detail != "" {
					errMsg += "\n\n" + d.Detail
				}

				return nil, errors.New(errMsg)
			}
		}
	}

	for _, s := range d.Get("setting").(*schema.Set).List() {
		delete(s.(map[string]interface{}), "description")
	}

	return []*schema.ResourceData{d}, nil
}

func resourceRdsConfigurationCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	config, err := readConfiguration(ctx, meta.(*mysqlClient), true)
	if err != nil {
		return err
	}

	badNames := []string{}

	for _, s := range d.Get("setting").(*schema.Set).List() {
		name := s.(map[string]interface{})["name"].(string)

		if _, ok := config[name]; !ok {
			badNames = append(badNames, name)
		}
	}

	if len(badNames) > 0 {
		errMsg := fmt.Sprintf(
			"Unsupported RDS configuration settings: \"%s\"\n\n"+
				"Valid settings are:\n",
			strings.Join(badNames, "\", \""),
		)

		for i, c := range maps.Values(config) {
			errMsg += fmt.Sprintf("% 3d. \"%s\": %s\n", (i + 1), c["name"], c["description"])
		}

		return errors.New(errMsg)
	}

	return nil
}
