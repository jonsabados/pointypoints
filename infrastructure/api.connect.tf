data "aws_iam_policy_document" "connect_lambda_policy" {
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
}

resource "aws_iam_role" "connect_lambda_role" {
  name = "connectLambdaRole"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role_policy.json

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_iam_role_policy" "connect_lambda_role_policy" {
  role = aws_iam_role.connect_lambda_role.name
  policy = data.aws_iam_policy_document.connect_lambda_policy.json
}

resource "aws_lambda_function" "connect_lambda" {
  filename = "../dist/connectLambda.zip"
  source_code_hash = filebase64sha256("../dist/connectLambda.zip")
  handler = "connect"
  function_name = "${local.workspace_prefix}connect"
  role = aws_iam_role.connect_lambda_role.arn
  runtime = "go1.x"

  tracing_config {
    mode = "Active"
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_cloudwatch_log_group" "connect_lambda_logs" {
  name = "/aws/lambda/${aws_lambda_function.connect_lambda.function_name}"
  retention_in_days = 7
}

resource "aws_apigatewayv2_integration" "connect_integration" {
  api_id = aws_apigatewayv2_api.api.id
  integration_type = "AWS"

  connection_type = "INTERNET"
  content_handling_strategy = "CONVERT_TO_TEXT"
  description = "Disconnect Lambda Integration"
  integration_method = "POST"
  integration_uri = aws_lambda_function.connect_lambda.invoke_arn
  passthrough_behavior = "WHEN_NO_MATCH"
}

resource "aws_apigatewayv2_route" "connect" {
  api_id = aws_apigatewayv2_api.api.id
  route_key = "$connect"
  target = "integrations/${aws_apigatewayv2_integration.connect_integration.id}"
}