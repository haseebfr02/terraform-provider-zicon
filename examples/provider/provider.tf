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
