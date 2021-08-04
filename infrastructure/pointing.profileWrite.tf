data "aws_iam_policy_document" "profile_write_lambda" {
  statement {
    sid    = "AllowLogging"
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
    sid    = "AllowXRayWrite"
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
    sid    = "AllowProfileAccess"
    effect = "Allow"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/${aws_dynamodb_table.profile_store.name}"
    ]
  }
}


module "profileWrite_lambda" {
  source = "./rest-endpoint"

  aws_region = var.aws_region
  api_id     = aws_api_gateway_rest_api.rest_pointing.id

  name   = "profileWrite"
  policy = data.aws_iam_policy_document.profile_write_lambda.json

  lambda_env = {
    LOG_LEVEL       = "info"
    ALLOWED_ORIGINS = "https://${module.ui_cert.distinct_domain_names[0]},https://${module.ui_cert.distinct_domain_names[1]},http://localhost:8080"
    "PROFILE_TABLE" : aws_dynamodb_table.profile_store.name,
  }

  http_method = "PUT"
  resource_id = aws_api_gateway_resource.profile.id
  full_path   = aws_api_gateway_resource.profile.path

  authorizer_id = aws_api_gateway_authorizer.authorizer.id

  request_parameters = {
  }
}