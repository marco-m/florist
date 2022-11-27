# 🌼 florist 🌺

A bare-bones and opinionated Go package to create **non idempotent**, one-file-contains-everything installers/provisioners.

## Status

**The project currently does not accept Pull Requests.**

**Work in progress, heavily unstable, not ready for use.**

## Goals

- Stay small and as simple as possible.
- Zero-dependencies (no interpreters, no additional files, the installer is all you need).
- Configuration-is-code, static typing (no YAML, less runtime errors, full programming language power).

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

We use VM snapshots, so that we can quickly restore a pristine environment.

Prepare a VM from scratch, provision and take a snapshot. You need to do this only once:

```
task vm:init
```

## Developing your installer/flowers

You can do all development on the VM, or use the VM only to run the installer and the tests and do the development and build on the host.
It is up to your convenience.

This is the normal sequence to perform when developing.

    # Restore snapshot and restart VM
    $ task vm:restore

    # Connect to the VM directly (bypass vagrant, way faster)
    $ ssh -F ssh.config.vagrant florist-dev

    == following happens in the VM ==

    # The florist/ directory is a shared mount with the host
    vagrant@florist-dev:~$ cd florist

    # Check out the example installer
    vagrant@florist-dev $ ./bin/example-florist -h

Then:

1. Develop and build on the host
2. Run the installer on the VM, check around
3. When the VM environment is dirty, restore the snapshot and go back to 1.
   The code is safe on the host (and editable also on the VM via the shared directory)

## Warning: testing your installer/flowers

By their nature, the flowers alter in a persistent way the global state of the machine: add/remove packages, add users, add/modify system files, add system services ...

As such, you don't want to run the installer on your development machine. Same reasoning for the majority of the tests.

The convention taken by the tests to reduce the possibility of an error is the following.

When preparing the VM, target `vm:init` will create directory `opt/florist/disposable`; its presence means that the machine can be subjected to destructive tests.

Tests that exercise a destructive functionality begin with

```go
func TestSshAddAuthorizedKeysVM(t *testing.T) {
    florist.SkipIfNotDisposableHost(t)
    ...
```

so that they will be skipped when running by mistake on the host. You can use the same conventions for your own tests.

Note that the flowers themselves are _not_ protected by the equivalent of `SkipIfNotDisposableHost`: you must NOT run the installer on your host, only on a test VM.

## Examples

See directory [examples/](examples) for example installers.

See section [Development](#development) for how to run the example installer in the VM.
