terraform {
  required_providers {
    zicon = {
      source = "haseebfr02/zicon"
    }
  }
}

variable "zicon_access_token" {
  description = "Supabase session token used to authenticate against the ZiCON Cloud API."
  type        = string
  sensitive   = true
}

provider "zicon" {
  access_token = var.zicon_access_token
}

resource "zicon_project" "example" {
  name        = "terraform-managed-project"
  category    = "Test"
  description = "created via terraform"
}

output "project_id" {
  value = zicon_project.example.id
}

output "project_api_key" {
  value     = zicon_project.example.api_key
  sensitive = true
}
