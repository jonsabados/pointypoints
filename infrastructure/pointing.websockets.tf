resource "aws_cloudwatch_log_group" "websockets_gateway_logs" {
  name              = "/aws/apigateway/${aws_apigatewayv2_api.websockets_pointing.id}/${aws_apigatewayv2_api.websockets_pointing.name}-main"
  retention_in_days = 7
}

module "websockets_pointing_cert" {
  source  = "terraform-aws-modules/acm/aws"
  version = "3.2.0"

  domain_name         = "${local.workspace_prefix}pointing-events.${data.aws_ssm_parameter.domain_name.value}"
  zone_id             = data.aws_route53_zone.main_domain.id
  wait_for_validation = true

  tags = {
    Name      = "${local.workspace_prefix}pointing-events.${data.aws_ssm_parameter.domain_name.value}"
    Workspace = terraform.workspace
  }
}

resource "aws_apigatewayv2_api" "websockets_pointing" {
  name                       = "${local.workspace_prefix}pointing-events"
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_apigatewayv2_deployment" "websockets_pointing" {
  api_id = aws_apigatewayv2_api.websockets_pointing.id

  triggers = {
    redeployment = sha1(jsonencode(concat(
      module.connect_lambda.change_keys,
      module.disconnect_lambda.change_keys,
      module.ping_lambda.change_keys,
    )))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_apigatewayv2_stage" "websockets_pointing_stage" {
  api_id        = aws_apigatewayv2_api.websockets_pointing.id
  name          = "${local.workspace_prefix}pointing-main"
  deployment_id = aws_apigatewayv2_deployment.websockets_pointing.id

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.websockets_gateway_logs.arn
    format          = "{ \"requestId\":\"$context.requestId\",\"ip\": \"$context.identity.sourceIp\", \"caller\":\"$context.identity.caller\", \"user\":\"$context.identity.user\", \"requestTime\":\"$context.requestTime\", \"httpMethod\":\"$context.httpMethod\", \"resourcePath\":\"$context.resourcePath\", \"status\":\"$context.status\", \"protocol\":\"$context.protocol\", \"responseLength\":\"$context.responseLength\" }"
  }

  default_route_settings {
    data_trace_enabled     = true
    logging_level          = "ERROR"
    throttling_burst_limit = 5000
    throttling_rate_limit  = 10000
  }
}

resource "aws_apigatewayv2_api_mapping" "websockets_pointing" {
  api_id      = aws_apigatewayv2_api.websockets_pointing.id
  stage       = aws_apigatewayv2_stage.websockets_pointing_stage.name
  domain_name = aws_apigatewayv2_domain_name.websockets_pointing.domain_name
}

resource "aws_apigatewayv2_domain_name" "websockets_pointing" {
  domain_name = "${local.workspace_prefix}pointing-events.${data.aws_ssm_parameter.domain_name.value}"

  domain_name_configuration {
    certificate_arn = module.websockets_pointing_cert.acm_certificate_arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_route53_record" "websockets_pointing" {
  name    = aws_apigatewayv2_domain_name.websockets_pointing.domain_name
  type    = "A"
  zone_id = data.aws_route53_zone.main_domain.zone_id

  alias {
    name                   = aws_apigatewayv2_domain_name.websockets_pointing.domain_name_configuration[0].target_domain_name
    zone_id                = aws_apigatewayv2_domain_name.websockets_pointing.domain_name_configuration[0].hosted_zone_id
    evaluate_target_health = false
  }
}
