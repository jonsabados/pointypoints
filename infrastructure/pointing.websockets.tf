resource "aws_cloudwatch_log_group" "websockets_gateway_logs" {
  name              = "/aws/apigateway/${aws_apigatewayv2_api.websockets_pointing.id}/${aws_apigatewayv2_api.websockets_pointing.name}-main"
  retention_in_days = 7
}

resource "aws_acm_certificate" "websockets_pointing_cert" {
  domain_name       = "${local.workspace_prefix}pointing-events.${data.aws_ssm_parameter.domain_name.value}"
  validation_method = "DNS"

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_route53_record" "pointing_cert_cert_verification_record" {
  count   = 1
  name    = aws_acm_certificate.websockets_pointing_cert.domain_validation_options[count.index].resource_record_name
  type    = aws_acm_certificate.websockets_pointing_cert.domain_validation_options[count.index].resource_record_type
  zone_id = data.aws_route53_zone.main_domain.id
  records = [aws_acm_certificate.websockets_pointing_cert.domain_validation_options[count.index].resource_record_value]
  ttl     = 300
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
  depends_on = [
    module.connect_lambda,
    module.disconnect_lambda,
    module.newSession_lambda,
    module.loadFacilitatorSession_lambda,
    module.loadSession_lambda,
    module.joinSession_lambda,
    module.vote_lambda,
    module.showVotes_lambda,
    module.clearVotes_lambda,
    module.ping_lambda
  ]

  api_id = aws_apigatewayv2_api.websockets_pointing.id

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
    certificate_arn = aws_acm_certificate.websockets_pointing_cert.arn
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
