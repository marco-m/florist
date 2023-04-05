# 🌼 florist 🌺

A bare-bones and opinionated Go package to create a **non idempotent**, one-file-contains-everything provisioner (install and configure).

## Status

**The project currently does not accept Pull Requests.**

**Work in progress, heavily unstable, not ready for use.**

## Goals

- Stay small and as simple as possible.
- Zero-dependencies (no interpreters, no additional files, the installer is all you need).
- Configuration-is-code, static typing (no YAML, less runtime errors, full programming language power).

## Use cases

- Installer: substitute shell scripts / Salt / Ansible / similar tools when building an image with Packer.
- Configurer: substitute cloud-init / shell scripts / Salt / Ansible / similar tools to configure an image at deployment time with Terraform.

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
    - use some ready-made flowers in this module.
- **bouquet** a target for the `install` subcommand, made of one or more flowers. You can list the installable bouquets with the `list` subcommand.

## The install and configure subcommands.

- Use `install` when building the image with Packer.
- Use `configure` when deploying the image with Terraform.

See also section [Secrets](#secrets) to understand Florist and secrets handling.

## Files: embed at compile time or download at runtime

Florist uses Go [embed](https://pkg.go.dev/embed) to recursively embed all files below a directory. The conventional name of the directory is `files/` (can be overridden by each flower). You will then pass along the `embed.FS` to the various flowers.

To see all the embedded files in a given installer, run it with the `embed-list` subcommand, or run `go list -f '{{.EmbedFiles}}'` in the directory containing the main package of the installer.

It is also possible to download files at runtime, using `florist.NetFetch` and then unarchive with `florist.UnzipOne`.

## Text templates

Florist supports [Go text templates] as follows.

- Each exported field of a `Flower` is available as template field.
- Template processing is done in one of the functions `florist.CopyFileTemplate`,`florist.CopyFileTemplateFromFs`. Just pass the flower as the tmplData parameter.

Since the template data is a struct (as opposed to a map), any template field error will result in an error.

See `os_test.go` for an example.

## Default values

Thanks to the [defaults package], you can set default values for flowers fields (pun intended!), with or without text templates:

```go
type Flower struct {
	FilesFS fs.FS
	Port    int `default:"22"`
	log     hclog.Logger
}
```

## Secrets

In general, do NOT store any secret on the image at image build time (`florist install`). Instead, inject secrets only in the running instance.

You have two options:
- Use `cloud-init` or equivalent.
- Use `florist configure`, as explained in the rest of this section.

Secrets handling in Florist depend on how deep you are bootstrapping your infrastructure:
- If at Terraform time you have available something on the network, such as a secrets store like Vault or SSM, or a KV store like Consul, use it (currently Florist doesn't have an API for this, but it is easy to add it yourself).
- If at Terraform time you do not have available anything on the network, you can embed the secrets in florist. This is safe, as long as:
  - the florist executable is purpose-built for the specific Terraform root or instance.
  - the florist executable is uploaded on the instance with a `file` provisioner and executed with a `remote-exec` provisioner in the same `null_resource`.
  - the florist executable is deleted form the instance, also in case of failure.
  - the florist executable, together with the files containing the secrets (that have been embedded), are deleted from the local host just after the `terraform apply` (also in case of failure!)
  - all this means that you need to drive all this sequence from a build script carefully written.

For a real-world example, see the orsolabs project (FIXME ADD LINK)


## Usage

```
$ ./example-florist -h
🌼 florist 🌺 - a simple provisioner

Usage: example-florist <command> [<args>]

Options:
  --help, -h             display this help and exit

Commands:
  install                install one or more bouquets
  configure              configure one or more bouquets
  list                   list the available bouquets
  embed-list             list the embedded FS
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


[Go text templates]: https://pkg.go.dev/text/template
[defaults package]:  https://github.com/creasty/defaults
