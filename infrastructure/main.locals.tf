locals {
  workspace_prefix        = terraform.workspace == "default" ? "" : "${terraform.workspace}-"
  workspace_domain_prefix = terraform.workspace == "default" ? "" : "${terraform.workspace}."
}