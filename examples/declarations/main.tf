terraform {
  required_providers {
    mergeas3 = {
      version = "0.1.0"
      source  = "euphario/f5/mergeas3"
    }
  }
}

variable "folder" {
  type = string
}

variable "schema_version" {
  type    = string
  default = "3.27.0"
}

variable "validation" {
  type    = bool
  default = true
}

variable "schema" {
  type    = string
  default = "https://raw.githubusercontent.com/F5Networks/f5-appsvcs-extension/master/schema/3.27.0/as3-schema.json"
}

data "mergeas3" "all" {
  folder         = var.folder
  schema_version = var.schema_version
  validation     = var.validation
}

output "declarations" {
  value = {
    for bigip in data.mergeas3.all.declarations :
    bigip.tag => bigip.as3
  }
}
