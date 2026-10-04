package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	// Choose a location near you.
	// https://docs.hetzner.com/cloud/general/locations/#what-locations-are-there
	Location = "nbg1"

	// The less expensive:
	// 2 vCPU, 4 GB RAM, 40 GB SSD, 20 TB traffic, €5.93/month
	ServerType = "cx23"

	OsVersion = "debian-12"
	// OsVersion = "debian-13"

	AnyIPv4 = pulumi.String("0.0.0.0/0")
	AnyIPv6 = pulumi.String("::/0")
)

// Treat this as a constant!
var kUsers = []string{"florist", "banana"}

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
			Name:             pulumi.StringRef(OsVersion),
			MostRecent:       pulumi.BoolRef(true),
			WithArchitecture: pulumi.StringRef("x86"),
		})
	if err != nil {
		return err
	}

	createHook, err := ctx.RegisterResourceHook("ssh-on-create", sshConfigCreateHook, nil)
	if err != nil {
		return err
	}
	deleteHook, err := ctx.RegisterResourceHook("ssh-on-delete", sshConfigDeleteHook, nil)
	if err != nil {
		return err
	}

	// https://docs.hetzner.com/cloud/servers/overview#resources-and-attributes
	server, err := hcloud.NewServer(ctx, "florist",
		&hcloud.ServerArgs{
			// https://www.hetzner.com/cloud/#pricing
			ServerType: pulumi.String(ServerType),
			Image:      pulumi.String(strconv.Itoa(*image.Id)),
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
			// Set the SSH public keys for the root user to the public keys with these
			// names. Must have been provisioned beforehand in the Hetzner console for
			// the project.
			SshKeys: pulumi.ToStringArray(kUsers),
		},
		pulumi.ResourceHooks(&pulumi.ResourceHookBinding{
			AfterCreate: []*pulumi.ResourceHook{createHook},
			AfterDelete: []*pulumi.ResourceHook{deleteHook},
		}),
	)
	if err != nil {
		return err
	}
	ctx.Export("IPv4", server.Ipv4Address)
	ctx.Export("IPv6", server.Ipv6Address)
	ctx.Export("name", server.Name)

	return nil
}

// ctx.RootDirectory()
const sshConfigFileName = "../ssh.config.pulumi"

func sshConfigCreateHook(args *pulumi.ResourceHookArgs) error {
	// NOTE this is a create hook:
	// args.OldIputs, args.OldOutputs are empty
	// args.NewIputs, args.NewOutputs are full

	ipv4 := args.NewOutputs["ipv4Address"].StringValue()
	ipv6 := args.NewOutputs["ipv6Address"].StringValue()
	name := args.NewOutputs["name"].StringValue()

	fileName, err := filepath.Abs(sshConfigFileName)
	if err != nil {
		return err
	}
	fmt.Println("WRITING SSH CONFIG FILE", fileName)
	return writeSSHConfig(fileName, name, ipv4, ipv6)
}

func sshConfigDeleteHook(args *pulumi.ResourceHookArgs) error {
	// NOTE this is a delete hook:
	// args.OldIputs, args.OldOutputs are full
	// args.NewIputs, args.NewOutputs are empty

	ipv4 := args.OldOutputs["ipv4Address"].StringValue()
	ipv6 := args.OldOutputs["ipv6Address"].StringValue()
	name := args.OldOutputs["name"].StringValue()

	fmt.Println("old IPv4", ipv4)
	fmt.Println("old IPv6", ipv6)
	fmt.Println("old serverName", name)

	fileName, err := filepath.Abs(sshConfigFileName)
	if err != nil {
		return err
	}
	fmt.Println("DELETING SSH CONFIG FILE", fileName)
	// If an after hook returns an error, Pulumi will log a warning diagnostic and
	// the Pulumi operation will continue, which is what we want in this case.
	// See https://www.pulumi.com/docs/iac/concepts/options/hooks/
	// UPDATE I saw that returning an error now misses up pulumi ???
	// return os.Remove(fileName)
	if err := os.Remove(fileName); err != nil {
		fmt.Println(err)
	}
	return nil
}

func writeSSHConfig(fileName, serverName, ipv4Address, ipv6Address string) error {
	out, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("createSSHConfig: %w", err)
	}

	_, err = fmt.Fprintf(out, `
# Generated by pulumi/main.go. DO NOT EDIT.
#
# USAGE: Add at the top (must be at the top) of $HOME/.ssh/config the line:
#   Include ~/src/hw/florist/ssh.config.pulumi
#
# Then, you can simply do "ssh florist-hcloud" without the need to specify the
# flag "-F ssh.config.pulumi", nor to ssh nor to any utility that uses ssh.

Host florist-hcloud
  # (Pulumi: %s)
  User root
  # Mutagen doesn't like IPv6 ???
  #HostName %s
  HostName %s
  CheckHostIP no
  StrictHostKeyChecking no
  #UserKnownHostsFile ~/.ssh/florist-known_hosts
  UserKnownHostsFile /dev/null
  IdentityFile ~/.ssh/florist.pub
  IdentitiesOnly yes
  RequestTTY yes
`, serverName, ipv6Address, ipv4Address)
	if err != nil {
		return fmt.Errorf("createSSHConfig: %w", err)
	}
	return nil
}
