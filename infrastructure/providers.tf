variable "aws_region" {
  type    = string
  default = "us-east-1"
}

provider "aws" {
  region  = var.aws_region
}