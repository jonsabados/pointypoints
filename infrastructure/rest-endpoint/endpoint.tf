data "aws_caller_identity" "current" {}

locals {
  workspace_prefix = terraform.workspace == "default" ? "" : "${terraform.workspace}-"
}

data "aws_iam_policy_document" "assume_lambda_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = [
        "lambda.amazonaws.com"
      ]

      type = "Service"
    }

    effect = "Allow"
    sid    = "AllowLambdaAssumeRole"
  }
}

resource "aws_iam_role" "lambda_role" {
  name               = "${local.workspace_prefix}${var.name}PointingLambdaRole"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role_policy.json

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_iam_role_policy" "lambda_role_policy" {
  role   = aws_iam_role.lambda_role.name
  policy = var.policy
}

resource "aws_lambda_function" "lambda" {
  filename         = "../dist/${var.name}Lambda.zip"
  source_code_hash = filebase64sha256("../dist/${var.name}Lambda.zip")
  handler          = var.name
  function_name    = "${local.workspace_prefix}${var.name}"
  role             = aws_iam_role.lambda_role.arn
  runtime          = "go1.x"

  tracing_config {
    mode = "Active"
  }

  environment {
    variables = var.lambda_env
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${aws_lambda_function.lambda.function_name}"
  retention_in_days = 7
}

resource "aws_api_gateway_method" "method" {
  authorization = "NONE"
  http_method   = var.http_method
  resource_id   = var.resource_id
  rest_api_id   = var.api_id

  request_parameters = var.request_parameters
}

resource "aws_api_gateway_integration" "integration" {
  rest_api_id             = var.api_id
  resource_id             = var.resource_id
  http_method             = aws_api_gateway_method.method.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.lambda.invoke_arn
}

resource "aws_lambda_permission" "allow_gateway_invoke" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "arn:aws:execute-api:us-east-1:${data.aws_caller_identity.current.account_id}:${var.api_id}/*/${var.http_method}${var.full_path}"
}