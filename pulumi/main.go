package main

import (
	"strconv"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	// Choose one close to you.
	// https://docs.hetzner.com/cloud/general/locations/#what-locations-are-there
	Location = "nbg1"

	AnyIPv4 = pulumi.String("0.0.0.0/0")
	AnyIPv6 = pulumi.String("::/0")

	User = "florist"
)

func main() {
	pulumi.Run(run)
}

func run(ctx *pulumi.Context) error {
	labels := pulumi.ToStringMap(map[string]string{
		"project": ctx.Project(),
		"env":     ctx.Stack(),
	})

	firewall, err := hcloud.NewFirewall(ctx, "firewall",
		&hcloud.FirewallArgs{
			Labels: labels,
			Rules: hcloud.FirewallRuleArray{
				&hcloud.FirewallRuleArgs{
					Description: pulumi.String("inbound ICMP ping"),
					Direction:   pulumi.String("in"),
					Protocol:    pulumi.String("icmp"),
					SourceIps:   pulumi.StringArray{AnyIPv4, AnyIPv6},
				},
				&hcloud.FirewallRuleArgs{
					Description: pulumi.String("inbound SSH"),
					Direction:   pulumi.String("in"),
					Protocol:    pulumi.String("tcp"),
					Port:        pulumi.String("22"),
					SourceIps:   pulumi.StringArray{AnyIPv4, AnyIPv6},
				},
			},
		})
	if err != nil {
		return err
	}

	image, err := hcloud.GetImage(ctx,
		&hcloud.GetImageArgs{
			Name:             pulumi.StringRef("debian-12"),
			MostRecent:       pulumi.BoolRef(true),
			WithArchitecture: pulumi.StringRef("x86"),
		})
	if err != nil {
		return err
	}

	// https://docs.hetzner.com/cloud/servers/overview#resources-and-attributes
	server, err := hcloud.NewServer(ctx, "florist", &hcloud.ServerArgs{
		// https://www.hetzner.com/cloud/#pricing
		ServerType: pulumi.String("cx22"),
		Image:      pulumi.String(strconv.Itoa(image.Id)),
		Location:   pulumi.String(Location),
		Labels:     labels,
		PublicNets: hcloud.ServerPublicNetArray{
			&hcloud.ServerPublicNetArgs{
				// Public IPv4, extra cost. Enable if needed.
				// Must enable to be able to reach among others github.com :-(
				// https://docs.hetzner.com/cloud/servers/primary-ips/overview#pricing
				Ipv4Enabled: pulumi.BoolPtr(true),
				// Public IPv6, no extra cost.
				Ipv6Enabled: pulumi.BoolPtr(true),
			},
		},
		FirewallIds: pulumi.IntArray{firewall.ID().ApplyT(strconv.Atoi).(pulumi.IntOutput)},
		// Set the SSH public key for the root user to the key with this name.
		SshKeys: pulumi.ToStringArray([]string{User}),
	})
	if err != nil {
		return err
	}
	ctx.Export("IPv4", server.Ipv4Address)
	ctx.Export("IPv6", server.Ipv6Address)
	ctx.Export("name", server.Name)

	return nil
}
