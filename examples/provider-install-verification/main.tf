terraform {
  required_providers {
    webarena = {
      source = "registry.terraform.io/kunitsucom/webarena"
    }
  }
}

provider "webarena" {}

data "webarena_indigo_v1_vm_sshkey" "example" {
    id = 21750
}
