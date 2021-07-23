data "aws_iam_policy_document" "session_modifying_lambda_policy" {
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
    sid    = "AllowSessionAccess"
    effect = "Allow"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:Query",
      "dynamodb:DeleteItem",
      "dynamodb:PutItem",
      "dynamodb:DescribeStream",
      "dynamodb:DescribeTable"
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/${aws_dynamodb_table.session_store.name}"
    ]
  }

  statement {
    sid    = "AllowSessionSocketIndexQuery"
    effect = "Allow"
    actions = [
      "dynamodb:Query"
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/${aws_dynamodb_table.session_store.name}/index/${local.session_socket_index_name}"
    ]
  }

  statement {
    sid    = "AllowMessages"
    effect = "Allow"
    actions = [
      "execute-api:ManageConnections"
    ]
    resources = [
      "arn:aws:execute-api:${var.aws_region}:${data.aws_caller_identity.current.account_id}:${aws_apigatewayv2_api.websockets_pointing.id}/*"
    ]
  }
}

locals {
  session_modifying_lambda_env = {
    REGION               = var.aws_region
    GATEWAY_ENDPOINT     = "https://${aws_apigatewayv2_api.websockets_pointing.id}.execute-api.${var.aws_region}.amazonaws.com/${local.workspace_prefix}pointing-main/"
    SESSION_TABLE        = aws_dynamodb_table.session_store.name
    SESSION_SOCKET_INDEX = local.session_socket_index_name
    LOG_LEVEL            = "info"
    ALLOWED_ORIGINS      = "https://${aws_acm_certificate.ui_cert.domain_name},https://${aws_acm_certificate.ui_cert.subject_alternative_names[0]},http://localhost:8080"
  }
}