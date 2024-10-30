###########################
#### DataDog Forwarder ####
###########################

module "datadog_forwarder" {
  source = "../datadog_forwarder"

  datadog_api_key = var.datadog_api_key
  env             = terraform.workspace

  providers = {
    aws = aws
  }
}
