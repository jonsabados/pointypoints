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

module "connect_lambda" {
  source = "./endpoint"

  aws_region = var.aws_region

  api_id = aws_apigatewayv2_api.pointing.id
  name   = "connect"
  route  = "$connect"

  policy = data.aws_iam_policy_document.connect_lambda_policy.json

  lambda_env = {
    LOG_LEVEL = "info"
  }
}
