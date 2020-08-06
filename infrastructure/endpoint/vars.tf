variable "aws_region" {
  type = string
}

variable "policy" {
  type = string
}

variable "name" {
  type = string
}

variable "lambda_env" {
  type = map(string)
}

variable "api_id" {
  type = string
}

variable "route" {
  type = string
}