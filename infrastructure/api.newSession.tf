data "aws_iam_policy_document" "newSession_lambda_policy" {
  statement {
    sid = "AllowLogging"
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:aws:logs:*:*:*"
    ]
  }

  statement {
    sid = "AllowXRayWrite"
    effect = "Allow"
    actions = [
      "xray:PutTraceSegments",
      "xray:PutTelemetryRecords",
      "xray:GetSamplingRules",
      "xray:GetSamplingTargets",
      "xray:GetSamplingStatisticSummaries"
    ]
    resources = [
      "*"
    ]
  }

  statement {
    sid = "AllowSessionStart"
    effect = "Allow"
    actions = [
      "dynamodb:PutItem",
      "dynamodb:DescribeStream",
      "dynamodb:DescribeTable"
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/${aws_dynamodb_table.session_store.name}"
    ]
  }
}

resource "aws_iam_role" "newSession_lambda_role" {
  name = "newSessionLambdaRole"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role_policy.json

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_iam_role_policy" "newSession_lambda_role_policy" {
  role = aws_iam_role.newSession_lambda_role.name
  policy = data.aws_iam_policy_document.newSession_lambda_policy.json
}

resource "aws_lambda_function" "newSession_lambda" {
  filename = "../dist/newSessionLambda.zip"
  source_code_hash = filebase64sha256("../dist/newSessionLambda.zip")
  handler = "newSession"
  function_name = "${local.workspace_prefix}newSession"
  role = aws_iam_role.newSession_lambda_role.arn
  runtime = "go1.x"

  tracing_config {
    mode = "Active"
  }

  environment {
    variables = {
      SESSION_TABLE = aws_dynamodb_table.session_store.name
    }
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_cloudwatch_log_group" "newSession_lambda_logs" {
  name = "/aws/lambda/${aws_lambda_function.newSession_lambda.function_name}"
  retention_in_days = 7
}

resource "aws_apigatewayv2_integration" "newSession_integration" {
  api_id = aws_apigatewayv2_api.api.id
  integration_type = "AWS"

  connection_type = "INTERNET"
  content_handling_strategy = "CONVERT_TO_TEXT"
  description = "New Session Lambda Integration"
  integration_method = "POST"
  integration_uri = aws_lambda_function.newSession_lambda.invoke_arn
  passthrough_behavior = "WHEN_NO_MATCH"
}

resource "aws_apigatewayv2_route" "newSession" {
  api_id = aws_apigatewayv2_api.api.id
  route_key = "newSession"
  target = "integrations/${aws_apigatewayv2_integration.newSession_integration.id}"
}