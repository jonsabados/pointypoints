data "aws_iam_policy_document" "disconnect_lambda_policy" {
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

resource "aws_iam_role" "disconnect_lambda_role" {
  name = "disconnectLambdaRole"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role_policy.json

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_iam_role_policy" "disconnect_lambda_role_policy" {
  role = aws_iam_role.disconnect_lambda_role.name
  policy = data.aws_iam_policy_document.disconnect_lambda_policy.json
}

resource "aws_lambda_function" "disconnect_lambda" {
  filename = "../dist/disconnectLambda.zip"
  source_code_hash = filebase64sha256("../dist/disconnectLambda.zip")
  handler = "disconnect"
  function_name = "${local.workspace_prefix}disconnect"
  role = aws_iam_role.disconnect_lambda_role.arn
  runtime = "go1.x"

  tracing_config {
    mode = "Active"
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_cloudwatch_log_group" "disconnect_lambda_logs" {
  name = "/aws/lambda/${aws_lambda_function.disconnect_lambda.function_name}"
  retention_in_days = 7
}

resource "aws_apigatewayv2_integration" "disconnect_integration" {
  api_id = aws_apigatewayv2_api.api.id
  integration_type = "AWS"

  connection_type = "INTERNET"
  content_handling_strategy = "CONVERT_TO_TEXT"
  description = "Disconnect Lambda Integration"
  integration_method = "POST"
  integration_uri = aws_lambda_function.disconnect_lambda.invoke_arn
  passthrough_behavior = "WHEN_NO_MATCH"
}

resource "aws_apigatewayv2_route" "disconnect" {
  api_id = aws_apigatewayv2_api.api.id
  route_key = "$disconnect"
  target = "integrations/${aws_apigatewayv2_integration.disconnect_integration.id}"
}