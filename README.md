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
