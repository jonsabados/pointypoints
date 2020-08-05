data "aws_iam_policy_document" "joinSession_lambda_policy" {
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

  statement {
    sid       = "AllowSessionAccess"
    effect    = "Allow"
    actions   = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:DescribeStream",
      "dynamodb:DescribeTable"
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/${aws_dynamodb_table.session_store.name}"
    ]
  }

  statement {
    sid       = "AllowReadWatchers"
    effect    = "Allow"
    actions   = [
      "dynamodb:GetItem",
      "dynamodb:DescribeStream",
      "dynamodb:DescribeTable"
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/${aws_dynamodb_table.session_watcher_store.name}"
    ]
  }

  statement {
    sid       = "AllowLock"
    effect    = "Allow"
    actions   = [
      "dynamodb:PutItem",
      "dynamodb:GetItem",
      "dynamodb:DeleteItem",
      "dynamodb:DescribeStream",
      "dynamodb:DescribeTable"
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/${aws_dynamodb_table.global_locks.name}"
    ]
  }

  statement {
    sid       = "AllowMessages"
    effect    = "Allow"
    actions   = [
      "execute-api:ManageConnections"
    ]
    resources = [
      "arn:aws:execute-api:${var.aws_region}:${data.aws_caller_identity.current.account_id}:${aws_apigatewayv2_api.pointing.id}/*"
    ]
  }
}

resource "aws_iam_role" "joinSession_lambda_role" {
  name               = "${local.workspace_prefix}joinSessionLambdaRole"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role_policy.json

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_iam_role_policy" "joinSession_lambda_role_policy" {
  role   = aws_iam_role.joinSession_lambda_role.name
  policy = data.aws_iam_policy_document.joinSession_lambda_policy.json
}

resource "aws_lambda_function" "joinSession_lambda" {
  filename         = "../dist/joinSessionLambda.zip"
  source_code_hash = filebase64sha256("../dist/joinSessionLambda.zip")
  handler          = "joinSession"
  function_name    = "${local.workspace_prefix}joinSession"
  role             = aws_iam_role.joinSession_lambda_role.arn
  runtime          = "go1.x"

  tracing_config {
    mode = "Active"
  }

  environment {
    variables = {
      REGION           = var.aws_region
      GATEWAY_ENDPOINT = "https://${aws_apigatewayv2_api.pointing.id}.execute-api.${var.aws_region}.amazonaws.com/${local.workspace_prefix}pointing-main/"
      SESSION_TABLE    = aws_dynamodb_table.session_store.name
      WATCHER_TABLE    = aws_dynamodb_table.session_watcher_store.name
      LOCK_TABLE       = aws_dynamodb_table.global_locks.name
      LOG_LEVEL        = "info"
    }
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_cloudwatch_log_group" "joinSession_lambda_logs" {
  name              = "/aws/lambda/${aws_lambda_function.joinSession_lambda.function_name}"
  retention_in_days = 7
}

resource "aws_apigatewayv2_integration" "joinSession_integration" {
  api_id           = aws_apigatewayv2_api.pointing.id
  integration_type = "AWS_PROXY"

  description               = "Load Session Lambda Integration"
  integration_method        = "POST"
  integration_uri           = aws_lambda_function.joinSession_lambda.invoke_arn
  content_handling_strategy = "CONVERT_TO_TEXT"
  request_templates         = {}
}

resource "aws_apigatewayv2_route" "joinSession" {
  api_id    = aws_apigatewayv2_api.pointing.id
  route_key = "joinSession"
  target    = "integrations/${aws_apigatewayv2_integration.joinSession_integration.id}"
}

resource "aws_lambda_permission" "joinSession_allow_gateway_invoke" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.joinSession_lambda.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "arn:aws:execute-api:${var.aws_region}:${data.aws_caller_identity.current.account_id}:${aws_apigatewayv2_api.pointing.id}/*/joinSession"
}