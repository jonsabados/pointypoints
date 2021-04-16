data "aws_iam_policy_document" "ping_lambda_policy" {
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
    sid       = "AllowMessages"
    effect    = "Allow"
    actions   = [
      "execute-api:ManageConnections"
    ]
    resources = [
      "arn:aws:execute-api:${var.aws_region}:${data.aws_caller_identity.current.account_id}:${aws_apigatewayv2_api.websockets_pointing.id}/*"
    ]
  }
}

module "ping_lambda" {
  source = "./endpoint"

  aws_region = var.aws_region

  api_id = aws_apigatewayv2_api.websockets_pointing.id
  name   = "ping"
  route  = "ping"

  policy = data.aws_iam_policy_document.ping_lambda_policy.json

  lambda_env = {
    LOG_LEVEL        = "info"
    REGION           = var.aws_region
    GATEWAY_ENDPOINT = "https://${aws_apigatewayv2_api.websockets_pointing.id}.execute-api.${var.aws_region}.amazonaws.com/${local.workspace_prefix}pointing-main/"
  }
}
