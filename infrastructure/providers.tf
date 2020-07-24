variable "aws_region" {
  type    = string
  default = "us-east-1"
}

provider "aws" {
  version = "~> 2.8"
  region  = var.aws_region
}