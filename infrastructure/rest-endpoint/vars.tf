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

variable "full_path" {
  type = string
}

variable "resource_id" {
  type = string
}

variable "http_method" {
  type = string
}

variable "request_parameters" {
  type = map(string)
}
