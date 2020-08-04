data "aws_iam_policy_document" "connect_lambda_policy" {
  statement {
    sid       = "AllowLogging"
    effect    = "Allow"
    actions   = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:aws:logs:*:*:*"
    ]
  }

  statement {
    sid       = "AllowXRayWrite"
    effect    = "Allow"
    actions   = [
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
  name               = "${local.workspace_prefix}connectLambdaRole"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role_policy.json

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_iam_role_policy" "connect_lambda_role_policy" {
  role   = aws_iam_role.connect_lambda_role.name
  policy = data.aws_iam_policy_document.connect_lambda_policy.json
}

resource "aws_lambda_function" "connect_lambda" {
  filename         = "../dist/connectLambda.zip"
  source_code_hash = filebase64sha256("../dist/connectLambda.zip")
  handler          = "connect"
  function_name    = "${local.workspace_prefix}connect"
  role             = aws_iam_role.connect_lambda_role.arn
  runtime          = "go1.x"

  tracing_config {
    mode = "Active"
  }

  environment {
    variables = {
      LOG_LEVEL = "info"
    }
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_cloudwatch_log_group" "connect_lambda_logs" {
  name              = "/aws/lambda/${aws_lambda_function.connect_lambda.function_name}"
  retention_in_days = 7
}

resource "aws_apigatewayv2_integration" "connect_integration" {
  api_id           = aws_apigatewayv2_api.pointing.id
  integration_type = "AWS_PROXY"

  description               = "Connect Lambda Integration"
  integration_method        = "POST"
  integration_uri           = aws_lambda_function.connect_lambda.invoke_arn
  content_handling_strategy = "CONVERT_TO_TEXT"
  request_templates         = {}
}

resource "aws_apigatewayv2_route" "connect" {
  api_id                              = aws_apigatewayv2_api.pointing.id
  route_key                           = "$connect"
  target                              = "integrations/${aws_apigatewayv2_integration.connect_integration.id}"
  route_response_selection_expression = "$default"
}

resource "aws_lambda_permission" "connect_allow_gateway_invoke" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.connect_lambda.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "arn:aws:execute-api:us-east-1:${data.aws_caller_identity.current.account_id}:${aws_apigatewayv2_api.pointing.id}/*/$connect"
}