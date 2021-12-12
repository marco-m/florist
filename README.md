# 🌼 florist 🌺

A bare-bones and opinionated Go module to create **non idempotent**, one-file-contains-everything installers/provisioners.

## Status

**The project currently does not accept Pull Requests.**

**Work in progress, heavily unstable, not ready for use.**

## Goals

- Stay small and as simple as possible.
- Zero-dependencies (no interpreters, the installer is all you need).
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
- **flower** an installable unit, as a Go package, that implements the `flower` interface. You can:
  - write it for your project.
  - use 3rd-party flowers (they are just Go packages).
  - use some of the ready-made flowers in this module.

## Files: embed at compile time or download at runtime

Florist expects your code to use Go [embed](https://pkg.go.dev/embed) to recursively embed all files below directory `files/` (must be this name). You will then pass along the embed.FS to the various flowers.

It is also possible to download files at runtime.

To see all the embedded files, run `go list -f '{{.EmbedFiles}}'`.

## Templating

I would like to support Go templating in the configuration files, but it is not there yet.

## Secrets

I recommend to use another mechanism to inject secrets in your image and to inject them at deployment time: backing secrets in the image at Packer build time should be avoided. For example, use cloud-init driven from Terraform.

As a last and insecure resort, you can embed secrets in the installer, using by convention the `files/secrets/` directory. Do not commit the secrets in git.

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
    inline = ["sudo /tmp/<project>-florist install <FLOWER>"]
  }
}
```

## Examples

See directory [examples/](examples) for example installers.

### Running the examples in a VM

```text
$ vagrant up
$ vagrant ssh

== following happens in the VM ==

vagrant@florist-dev $ ./florist/bin/example-florist -h
...
```
