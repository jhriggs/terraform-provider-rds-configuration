package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"golang.org/x/exp/maps"
)

func dataSourceRdsConfiguration() *schema.Resource {
	return &schema.Resource{
		Description: "RDS MySQL configuration settings.",
		ReadContext: dataSourceRdsConfigurationRead,

		Schema: map[string]*schema.Schema{
			"setting": {
				Description: "An individual RDS MySQL configuration setting.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "The name of the setting.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"value": {
							Description: "The value of the setting.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"description": {
							Description: "A description of the setting.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceRdsConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config, err := readConfiguration(ctx, meta.(*mysqlClient), true)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("setting", maps.Values(config))
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
