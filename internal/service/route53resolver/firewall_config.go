package route53resolver

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceFirewallConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFirewallConfigCreate,
		ReadWithoutTimeout:   resourceFirewallConfigRead,
		UpdateWithoutTimeout: resourceFirewallConfigUpdate,
		DeleteWithoutTimeout: resourceFirewallConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"firewall_fail_open": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(route53resolver.FirewallFailOpenStatus_Values(), false),
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceFirewallConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	input := &route53resolver.UpdateFirewallConfigInput{
		ResourceId: aws.String(d.Get("resource_id").(string)),
	}

	if v, ok := d.GetOk("firewall_fail_open"); ok {
		input.FirewallFailOpen = aws.String(v.(string))
	}

	output, err := conn.UpdateFirewallConfigWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Route53 Resolver Firewall Config: %s", err)
	}

	d.SetId(aws.StringValue(output.FirewallConfig.Id))

	return resourceFirewallConfigRead(ctx, d, meta)
}

func resourceFirewallConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	firewallConfig, err := FindFirewallConfigByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Firewall Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Route53 Resolver Firewall Config (%s): %s", d.Id(), err)
	}

	d.Set("firewall_fail_open", firewallConfig.FirewallFailOpen)
	d.Set("owner_id", firewallConfig.OwnerId)
	d.Set("resource_id", firewallConfig.ResourceId)

	return nil
}

func resourceFirewallConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	input := &route53resolver.UpdateFirewallConfigInput{
		ResourceId: aws.String(d.Get("resource_id").(string)),
	}

	if v, ok := d.GetOk("firewall_fail_open"); ok {
		input.FirewallFailOpen = aws.String(v.(string))
	}

	_, err := conn.UpdateFirewallConfigWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating Route53 Resolver Firewall Config: %s", err)
	}

	return resourceFirewallConfigRead(ctx, d, meta)
}

func resourceFirewallConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	log.Printf("[DEBUG] Deleting Route53 Resolver Firewall Config: %s", d.Id())
	_, err := conn.UpdateFirewallConfigWithContext(ctx, &route53resolver.UpdateFirewallConfigInput{
		ResourceId:       aws.String(d.Get("resource_id").(string)),
		FirewallFailOpen: aws.String(route53resolver.FirewallFailOpenStatusDisabled),
	})

	if err != nil {
		return diag.Errorf("deleting Route53 Resolver Firewall Config (%s): %s", d.Id(), err)
	}

	return nil
}

func FindFirewallConfigByID(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.FirewallConfig, error) {
	input := &route53resolver.ListFirewallConfigsInput{}
	var output *route53resolver.FirewallConfig

	// GetFirewallConfig does not support query by ID.
	err := conn.ListFirewallConfigsPagesWithContext(ctx, input, func(page *route53resolver.ListFirewallConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FirewallConfigs {
			if aws.StringValue(v.Id) == id {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{LastRequest: input}
	}

	return output, nil
}
