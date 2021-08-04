resource "aws_cloudwatch_log_group" "rest_gateway_logs" {
  name              = "/aws/apigateway/${aws_api_gateway_rest_api.rest_pointing.id}/${aws_api_gateway_rest_api.rest_pointing.name}-main"
  retention_in_days = 7
}

module "rest_pointing_cert" {
  source  = "terraform-aws-modules/acm/aws"
  version = "3.2.0"

  domain_name         = "${local.workspace_prefix}pointing.${data.aws_ssm_parameter.domain_name.value}"
  zone_id             = data.aws_route53_zone.main_domain.id
  wait_for_validation = true

  tags = {
    Name      = "${local.workspace_prefix}pointing.${data.aws_ssm_parameter.domain_name.value}"
    Workspace = terraform.workspace
  }
}

resource "aws_api_gateway_rest_api" "rest_pointing" {
  name = "${local.workspace_prefix}pointing"

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_api_gateway_domain_name" "rest_pointing" {
  domain_name     = "${local.workspace_prefix}pointing.${data.aws_ssm_parameter.domain_name.value}"
  certificate_arn = module.rest_pointing_cert.acm_certificate_arn

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_api_gateway_deployment" "rest_api" {
  rest_api_id = aws_api_gateway_rest_api.rest_pointing.id

  triggers = {
    redeployment = sha1(jsonencode(concat(
      module.cors_endpoint.change_keys,
      module.watchSession_lambda.change_keys,
      module.joinSession_lambda.change_keys,
      module.setFacilitatorSession_lambda.change_keys,
      module.vote_lambda.change_keys,
      module.updateSession_lambda.change_keys,
      module.newSession_lambda.change_keys,
      module.clearVotes_lambda.change_keys,
      module.profileRead_lambda.change_keys,
      module.profileWrite_lambda.change_keys,
    )))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "rest_pointing_main" {
  deployment_id = aws_api_gateway_deployment.rest_api.id
  rest_api_id   = aws_api_gateway_rest_api.rest_pointing.id
  stage_name    = "${local.workspace_prefix}rest-main"

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.rest_gateway_logs.arn
    format          = "{ \"requestId\":\"$context.requestId\",\"ip\": \"$context.identity.sourceIp\", \"caller\":\"$context.identity.caller\", \"user\":\"$context.identity.user\", \"requestTime\":\"$context.requestTime\", \"httpMethod\":\"$context.httpMethod\", \"resourcePath\":\"$context.resourcePath\", \"status\":\"$context.status\", \"protocol\":\"$context.protocol\", \"responseLength\":\"$context.responseLength\" }"
  }
}

resource "aws_route53_record" "pointing" {
  name    = "${local.workspace_prefix}pointing"
  type    = "CNAME"
  zone_id = data.aws_route53_zone.main_domain.id
  records = [aws_api_gateway_domain_name.rest_pointing.cloudfront_domain_name]
  ttl     = 300
}

resource "aws_api_gateway_base_path_mapping" "rest_pointing" {
  api_id      = aws_api_gateway_rest_api.rest_pointing.id
  stage_name  = aws_api_gateway_stage.rest_pointing_main.stage_name
  domain_name = aws_api_gateway_domain_name.rest_pointing.domain_name
}

resource "aws_api_gateway_resource" "session_path" {
  rest_api_id = aws_api_gateway_rest_api.rest_pointing.id
  parent_id   = aws_api_gateway_rest_api.rest_pointing.root_resource_id
  path_part   = "session"
}

resource "aws_api_gateway_resource" "session_var" {
  rest_api_id = aws_api_gateway_rest_api.rest_pointing.id
  parent_id   = aws_api_gateway_resource.session_path.id
  path_part   = "{session}"
}

resource "aws_api_gateway_resource" "user_path" {
  rest_api_id = aws_api_gateway_rest_api.rest_pointing.id
  parent_id   = aws_api_gateway_resource.session_var.id
  path_part   = "user"
}

resource "aws_api_gateway_resource" "user_var" {
  rest_api_id = aws_api_gateway_rest_api.rest_pointing.id
  parent_id   = aws_api_gateway_resource.user_path.id
  path_part   = "{user}"
}

data "aws_iam_policy_document" "cors_lambda_policy" {
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
    resources = ["*"]
  }
}

resource "aws_api_gateway_resource" "wildcard" {
  rest_api_id = aws_api_gateway_rest_api.rest_pointing.id
  parent_id   = aws_api_gateway_rest_api.rest_pointing.root_resource_id
  path_part   = "{proxy+}"
}

module "cors_endpoint" {
  source = "./rest-endpoint"

  aws_region = var.aws_region
  api_id     = aws_api_gateway_rest_api.rest_pointing.id

  name   = "cors"
  policy = data.aws_iam_policy_document.session_modifying_lambda_policy.json
  lambda_env = {
    LOG_LEVEL       = "info"
    ALLOWED_ORIGINS = "https://${module.ui_cert.distinct_domain_names[0]},https://${module.ui_cert.distinct_domain_names[1]},http://localhost:8080"
  }

  http_method = "OPTIONS"
  resource_id = aws_api_gateway_resource.wildcard.id
  full_path   = aws_api_gateway_resource.wildcard.path

  request_parameters = {}
}
