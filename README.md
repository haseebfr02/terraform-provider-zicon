# terraform-provider-zicon

A minimal Terraform provider for [ZiCON Cloud](https://bwwzefaddmyueynayize.supabase.co), built with
HashiCorp's [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).

It supports creating, reading, and deleting ZiCON Cloud projects (`zicon_project`). There is no
update endpoint on the API, so changing `name`, `category`, or `description` destroys and
recreates the project.

## Requirements

- [Go](https://go.dev/) 1.21+
- [Terraform](https://developer.hashicorp.com/terraform/downloads) 1.0+
- A ZiCON Cloud account and a Supabase session token (`access_token`) for it — project creation is
  authenticated with this session token, not a project API key.

## Building the provider

```sh
go build -o terraform-provider-zicon .
```

## Using the provider locally

Terraform providers are normally installed from a registry, but during development you can point
Terraform at your local build with a [CLI configuration override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers).

1. Build the binary as above.
2. Create (or edit) `~/.terraformrc`:

   ```hcl
   provider_installation {
     dev_overrides {
       "haseebfr02/zicon" = "/absolute/path/to/terraform-provider-zicon"
     }
     direct {}
   }
   ```

3. Run Terraform from the `examples/` directory (or your own config):

   ```sh
   cd examples
   export TF_VAR_zicon_access_token="<your Supabase session token>"
   terraform plan
   terraform apply
   ```

   With `dev_overrides` set, `terraform init` is not required (and will warn if you run it).

## Provider configuration

| Argument       | Required | Description                                                                          |
| -------------- | -------- | ------------------------------------------------------------------------------------- |
| `access_token` | Yes      | Supabase session token, used to authenticate `create-project` requests (sensitive).   |

## Resource: `zicon_project`

| Attribute     | Type   | Required/Computed          | Notes                                                            |
| ------------- | ------ | --------------------------- | ------------------------------------------------------------------ |
| `id`          | string | Computed                    | Assigned by the API on creation.                                  |
| `name`        | string | Required                    | Forces replacement on change (no update endpoint).                |
| `category`    | string | Optional                    | Forces replacement on change.                                     |
| `description` | string | Optional                    | Forces replacement on change.                                     |
| `api_key`     | string | Computed, sensitive         | Issued on creation; used for `list-projects` and `delete-project`. |

### Behavior

- **Create** — `POST create-project` with `Authorization: Bearer <access_token>` and a JSON body
  of `{name, category, description}`. The returned `id` and `api_key` are stored in state.
- **Read** — since there is no "get project by id" endpoint, `GET list-projects` is called with
  `X-ZiCON-Key: <api_key>` and the response is scanned for a matching `id`. A `401` (revoked key or
  deleted project) or a missing entry removes the resource from state so Terraform recreates it.
- **Delete** — `DELETE delete-project?id=<id>` with `X-ZiCON-Key: <api_key>`.
- **Update** — not supported by the API. `name`, `category`, and `description` all use
  `RequiresReplace`, so any change destroys and recreates the project.

## Documentation

Registry-facing documentation lives in [`docs/`](docs/) and is generated from the resource/provider
schema descriptions plus the example `.tf` files under [`examples/`](examples/) using
[`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs). The Terraform Registry requires
this `docs/` folder to be present and up to date at the tagged commit.

`tfplugindocs` is declared as a Go tool dependency (see the `tool` directive in `go.mod`), so no
separate install step is needed — regenerate the docs after changing any schema `Description` or
the files under `examples/provider/` and `examples/resources/<name>/`:

```sh
go generate ./...
```

Commit the resulting changes under `docs/` along with your schema change.

## Releasing

Releases are built and published automatically by [GoReleaser](https://goreleaser.com/) via the
[`.github/workflows/release.yml`](.github/workflows/release.yml) GitHub Actions workflow, following
[HashiCorp's provider release process](https://developer.hashicorp.com/terraform/registry/providers/publishing).

The workflow triggers whenever a tag matching `v*` (e.g. `v0.1.0`) is pushed to the repository:

```sh
git tag v0.1.0
git push origin v0.1.0
```

It then:

1. Checks out the full history (required so GoReleaser can generate a changelog/version info).
2. Builds binaries for `darwin`, `linux`, and `windows`, each for `amd64` and `arm64`.
3. Zips each binary and generates a `SHA256SUMS` checksum file, alongside the
   [`terraform-registry-manifest.json`](terraform-registry-manifest.json) manifest (protocol version 6,
   matching the Terraform Plugin Framework).
4. GPG-signs the checksum file (`SHA256SUMS.sig`) — the Terraform Registry requires this signature to
   verify release artifacts.
5. Publishes everything as a GitHub Release.

### Required GitHub secrets

The workflow needs a GPG key to sign release checksums. The corresponding public key
([`public-key.asc`](public-key.asc)) must also be added to your account/organization on the
[Terraform Registry](https://developer.hashicorp.com/terraform/registry/providers/publishing#generate-a-gpg-key)
so it can verify releases.

| Secret             | Description                                                                 |
| ------------------- | ---------------------------------------------------------------------------- |
| `GPG_PRIVATE_KEY`   | The ASCII-armored private key (`gpg --armor --export-secret-keys <key-id>`). |
| `PASSPHRASE`        | The passphrase protecting the private key.                                   |

Set these under **Settings → Secrets and variables → Actions** in the GitHub repository.
`GITHUB_TOKEN` is provided automatically by GitHub Actions and does not need to be configured.

## Example

See [`examples/main.tf`](examples/main.tf):

```hcl
terraform {
  required_providers {
    zicon = {
      source = "haseebfr02/zicon"
    }
  }
}

provider "zicon" {
  access_token = var.zicon_access_token
}

resource "zicon_project" "example" {
  name        = "terraform-managed-project"
  category    = "Test"
  description = "created via terraform"
}
```
