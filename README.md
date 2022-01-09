# 🌼 florist 🌺

A bare-bones and opinionated Go module to create **non idempotent**, one-file-contains-everything installers/provisioners.

## Status

**The project currently does not accept Pull Requests.**

**Work in progress, heavily unstable, not ready for use.**

## Goals

- Stay small and as simple as possible.
- Zero-dependencies (no interpreters, no additional files, the installer is all you need).
- Configuration-is-code, static typing (no YAML, less runtime errors).

## Use cases

- Substitute shell scripts / Salt / Ansible / similar tools to build images with tools like Packer.

## Non-goals

- Idempotency. If you need idempotency, use Ansible.
- Completeness. If you need completeness, use Ansible.
- Cover OSes or distributions that I don't use. Testing would become brittle and maintaining would become too time-consuming, I am sorry.

## Terminology

- **florist**: this module.
- **\<project\>-florist**: the installer for **\<project\>**. You write this one.
- **flower** a composable unit, under the form of a Go package, that implements the `flower` interface. You can:
    - write it for your project.
    - use 3rd-party flowers (they are just Go packages).
    - use some of the ready-made flowers in this module.
- **bouquet** a target for the `install` subcommand, made of one or more flowers. You can list the installable bouquets with the `list` subcommand.

## Files: embed at compile time or download at runtime

Florist expects your code to use Go [embed](https://pkg.go.dev/embed) to recursively embed all files below directory `files/` (must be this name). You will then pass along the embed.FS to the various flowers.

To see all the embedded files, run `go list -f '{{.EmbedFiles}}'`.

It is also possible to download files at runtime, using `florist.NetFetch()`.

## Templating

I would like to support Go templating in the configuration files, but it is not there yet.

## Secrets

I recommend to use another mechanism to inject secrets in your image and to inject them at deployment time: backing secrets in the image at Packer build time should be avoided. For example, use cloud-init driven from Terraform.

As a last and insecure resort, you can embed secrets in the installer, using by convention the `files/secrets/` directory. Do not commit the secrets in git.

## Usage

```
$ ./example-florist -h
🌼 florist 🌺 - a simple installer

Usage: example-florist <command> [<args>]

Options:
  --help, -h             display this help and exit

Commands:
  install                install one or more bouquets
  list                   list the available bouquets
```

## Usage with Packer

1. Build the installer.
2. Upload the installer with the `file` provisioner.
3. Run the installer with the `shell` provisioner.

Excerpt HCL configuration:

```HCL
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
```

## Prepare for development

We use VM snapshots, so that we can quickly restore a pristine environment:

    # Ensure we start from scratch
    $ vagrant destroy
    $ vagrant up

    # Take snapshot, name `pristine`
    $ vagrant snapshot save pristine

    # Generate a SSH configuration file
    $ vagrant ssh-config > ssh_config.vagrant

    # Edit ssh_config.vagrant and remove line `LogLevel FATAL`

## Developing your installer/flowers

You can do all development on the VM, or use the VM only to run the installer and the tests and do the development and build on the host.
It is up to your convenience.

This is the normal sequence to perform when developing.

    # Restore snapshot and restart VM
    $ vagrant snapshot restore pristine

    # Connect to the VM directly (bypass vagrant, way faster)
    $ ssh -F ssh_config.vagrant florist-dev

    == following happens in the VM ==

    # The florist/ directory is a shared mount with the host
    vagrant@florist-dev:~$ ls -F
    florist/

    # Check out the example installer
    vagrant@florist-dev $ ./florist/bin/example-florist -h

Then:

1. Develop and build on the host
2. Run the installer on the VM, check around
3. When the VM environment is dirty, restore the snapshot and go back to 1.
   The code is safe on the host (and editable also on the VM via the shared directory)

## Testing your installer/flowers

There are two ways to run the tests on the VM:

1. ssh to the VM and run the tests from a shell there as you would expect.
2. run the tests from the host using [xprog], a test runner for `go test -exec`.

Using `xprog` (see the [xprog] README for more information):

Assuming that the installer and its tests are in directory `\<project>-florist`:

From your host, cross-compile the tests and run them on the target VM 😀:

    $ GOOS=linux go test -exec="xprog ssh --cfg $PWD/ssh_config.vagrant --" ./<project>-florist

## Examples

See directory [examples/](examples) for example installers.

See section [Development](#development) for how to run the example installer in the VM.


[xprog]: https://github.com/marco-m/xprog