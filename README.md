# 🌼 florist 🌺

[![Go Reference](https://pkg.go.dev/badge/github.com/marco-m/florist.svg)](https://pkg.go.dev/github.com/marco-m/florist)
[![Build Status](https://api.cirrus-ci.com/github/marco-m/florist.svg?branch=master)](https://cirrus-ci.com/github/marco-m/florist)

A bare-bones and opinionated Go module to create a **non-idempotent**, **one-file-contains-everything** provisioner (install and configure).

## Status

**The project currently does not accept Pull Requests.**

**Work in progress, API heavily unstable, not ready for use.**

## Goals

- Stay small and as simple as possible.
- Zero-dependencies (no interpreters, no additional files, the provisioner executable is all you need).
- Configuration-is-code, static typing (no YAML, less runtime errors, full programming language power).

## Use cases

- Installer: substitute shell scripts / Salt / Ansible / similar tools when building an image with Packer.
- Configurer: substitute shell scripts / Salt / Ansible / similar tools to configure an image at deployment time with tools like cloud-init, Terraform, Pulumi.

## Non-goals

- Idempotency.
- Completeness.
- Mutable infrastructure / configuration management.
- Cover OSes or distributions that I don't use. Testing would become brittle and maintaining would become too time-consuming, I am sorry.

If you need any of these, then consider Ansible, SaltStack or similar.

## Terminology

- **florist**: this module.
- **\<role\>.florist**: the installer for **\<role\>**. You write one per role in your project.
- **flower** a composable unit, under the form of a Go package, that implements the [`flower`](./pkg/florist/florist.go) interface. You can:
  - write it for your project.
  - use 3rd-party flowers (they are just Go packages).
  - use some ready-made flowers in this module.

## The `install` and `configure` subcommands

- Use `install` when building the image with Packer or similar.
- Use `configure` when deploying the image with cloud-init, Terraform, Pulumi or similar.

## Files and templates: embed at compile time or download at runtime

Florist uses Go [embed](https://pkg.go.dev/embed) to recursively embed all files below a directory. The conventional name of the directory is `embedded` (can be overridden by each flower). You will then pass along the `embed.FS` to the various flowers.

To see all the embedded files in a given installer, run it with the `list` subcommand, or run `go list -f '{{.EmbedFiles}}'` in the directory containing the main package of the provisioner.

It is also possible to download files at runtime, using `florist.NetFetch` and then unarchive with `florist.UnzipOne`.

## Templating

Florist supports [Go text templates] with multiple functions:

- `TemplateFromText()`
- `TemplateFromFs()`
- `TemplateFromFsWithDelims()`. This replaces the default delimiters `{{`, `}}` with `<<` and `>>`. This is useful to reduce the clutter when the rendered file must contain `{{` or `}}`.

ee `os_test.go` for an example.

## Secrets

In general, do NOT store any secret on the image at image build time (`florist install`). Instead, inject secrets only in the running instance (`florist configure`).

Secrets handling in Florist depend on how deep you are bootstrapping your infrastructure. It can be roughly separated in 3 cases:

1. If at deployment time you have access to a network secrets manager (eg AWS Secrets Manager or SSM) and your infrastructure supports giving an identity to the instance (eg IAM roles) then you are all set. Just use the SDK of the secrets manager to fetch the secrets from the florist `configure` function.
2. If at deployment time you have access to a network secrets manager or K/V store (eg Vault, Consul, ...) but no instance identity, then pass an identity token to florist (see option 3) and then use the SDK of the secrets manager or K/V store to fetch the secrets from the florist `configure` function.
3. If at deployment time you do not have available anything on the network, then use an OOB method (eg: from Terraform, configure cloud-init to create a JSON file with the secrets, set it mode 06000) and invoke florist `configure --settings=secrets.json`.

For a real-world example, see the orsolabs project (FIXME ADD LINK)

## Usage

    $ ./example -h
    example -- A 🌼 florist 🌺 provisioner.
    Usage: example [--log-level LEVEL] <command> [<args>]
    
    Options:
      --log-level LEVEL      log level [default: INFO]
      --help, -h             display this help and exit
    
    Commands:
      list                   list the flowers and their files
      install
      configure

## Usage with Packer

1. Build the installer.
2. Upload the installer with the `file` provisioner.
3. Run the installer with the `shell` provisioner.

Excerpt HCL configuration:

    build {
      source "<provider>.cfg" {
        ...
      }
    
      provisioner "file" {
        source      = "path/to/<project>-florist"
        destination = "/tmp/<project>-florist"
      }
    
      provisioner "shell" {
        inline = ["sudo /tmp/<project>-florist install <BOUQUET>"]
      }
    }

## The development environment

By their nature, the flowers alter in a persistent way the global state of the target: add/remove packages, add users, add/modify system files, add system services ...

As such, you don't want to run the installer on your development machine. Same reasoning for the majority of the tests.

We cannot use Docker containers neither, because Florist expects an environment that is a poor match for containers. Classic example is the assumption that a system manager is present (eg: systemd), so that services can be added and configured.

Said in another way, a normal Florist target is bare-metal or a VM, not a container.

## Prepare for development

- Install [Vagrant](https://developer.hashicorp.com/vagrant/install).
- Install [VirtualBox](https://www.virtualbox.org/wiki/Downloads)
- Install QEMU and the Vagrant QEMU plugin. For macos: brew install qemu

    vagrant plugin install vagrant-qemu

We use VM snapshots, so that we can quickly restore a pristine environment.

Prepare a VM from scratch, provision and take a snapshot. You need to do this only once:

    task vm:init

## Prepare Pulumi and Hetzner Cloud

1. Install `hcloud`, the [Hetzner CLI](https://github.com/hetznercloud/cli).
2. In the [Hetzner console](https://console.hetzner.com), create a project, name it `florist`.
3. In the project, add the **public** key of your SSH key, name it `florist`.
4. In the project, create an API token and store it securely:

       envchain --set florist HCLOUD_TOKEN

5. Store the Pulumi passphrase in envchain:

       envchain --set florist PULUMI_CONFIG_PASSPHRASE

## Developing your installer/flowers

You can do all development on the VM, or use the VM only to run the installer and the tests and do the development and build on the host. It is up to your convenience.

This is the normal sequence to perform when developing.

    ## Restore snapshot and restart VM
    $ task vm:restore
    
    ## Connect to the VM directly (bypass vagrant, way faster)
    $ ssh -F ssh.config.vagrant florist-dev
    ... System installed by 🌼 florist 🌺
    
    ## == following happens in the VM ==
    
    ## The florist/ directory is a shared mount with the host
    vagrant@florist-dev:~$ cd florist
    
    ## Check out the example installer
    vagrant@florist-dev $ ./bin/example-florist -h

Then:

1. Develop and build on the host.
2. Run the installer on the VM, check around.
3. When the VM environment is **dirty**, restore the snapshot and go back to 1.
   The code is safe on the host (and editable also on the VM via the shared directory).

## Warning: testing your installer/flowers

By their nature, the flowers alter in a persistent way the global state of the machine: add/remove packages, add users, add/modify system files, add system services ...

As such, you don't want to run the installer on your development machine. Same reasoning for the majority of the tests.

The convention taken by the tests to reduce the possibility of an error is the following.

When preparing the VM, target `vm:init` will create directory `/opt/florist/disposable`; its presence means that the machine can be subjected to destructive tests.

Tests that exercise a destructive functionality begin with

    func TestSshAddAuthorizedKeysVM(t *testing.T) {
        florist.SkipIfNotDisposableHost(t)
        ...

so that they will be skipped when running by mistake on the host. You can use the same conventions for your own tests.

Note that the flowers themselves are _not_ protected by the equivalent of `SkipIfNotDisposableHost`: you must NOT run the installer on your host, only on a test VM.

## Testing

### From the host, run the tests on the guest

This will restore a pristine VM snapshot (very fast) and run the tests (leaving the VM running):

    task test:all:vm:clean

### From the guest, run the tests

    ## Connect to the VM directly (bypass vagrant, way faster)
    $ ssh -F ssh.config.vagrant florist-dev
    
    ## == following happens in the VM ==
    
    ## The florist/ directory is a shared mount with the host
    vagrant@florist-dev:~$ cd florist
    vagrant@florist-dev:~$ sudo task test:all

### Testing with a VM in Hetzner Cloud

See the Taskfile targets with prefix `pulumi`. Full cycle:

    task pulumi:all

### Useful files to troubleshoot

- The cloud-init logs are at `/var/log/cloud-init-output.log`
- The cloud-config file and other configuration information is at `/var/lib/cloud/instance/user-data.txt`

## Examples

See directory [examples/](examples) for example installers.

See section [Development](#prepare-for-development) for how to run the example installer in the VM.

[Go text templates]: https://pkg.go.dev/text/template
