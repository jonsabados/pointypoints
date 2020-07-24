data "aws_caller_identity" "current" {}

data "aws_ssm_parameter" "domain_name" {
  name = "pointypoints.domainname"
}

data "aws_route53_zone" "main_domain" {
  name = data.aws_ssm_parameter.domain_name.value
}

data "aws_iam_policy_document" "assume_lambda_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["lambda.amazonaws.com"]
      type        = "Service"
    }
    effect  = "Allow"
    sid     = "AllowLambdaAssumeRole"
  }
}

data "aws_iam_policy_document" "assume_gateway_role_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["apigateway.amazonaws.com"]
      type        = "Service"
    }
    effect  = "Allow"
    sid     = "AllowApiGatewayAssumeRole"
  }
}