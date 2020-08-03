resource "aws_acm_certificate" "api_cert" {
  domain_name = "${local.workspace_prefix}api.${data.aws_ssm_parameter.domain_name.value}"
  validation_method = "DNS"

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_route53_record" "api_cert_cert_verification_record" {
  count = 1
  name = aws_acm_certificate.api_cert.domain_validation_options[count.index].resource_record_name
  type = aws_acm_certificate.api_cert.domain_validation_options[count.index].resource_record_type
  zone_id = data.aws_route53_zone.main_domain.id
  records = [
    aws_acm_certificate.api_cert.domain_validation_options[count.index].resource_record_value]
  ttl = 300
}

resource "aws_apigatewayv2_api" "api" {
  name = "${local.workspace_prefix}api"
  protocol_type = "WEBSOCKET"
  route_selection_expression = "$request.body.action"

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_apigatewayv2_deployment" "main" {
  depends_on = [
    aws_apigatewayv2_integration.newSession_integration,
    aws_apigatewayv2_integration.disconnect_integration
  ]
  api_id = aws_apigatewayv2_api.api.id

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_apigatewayv2_stage" "api_stage" {
  api_id = aws_apigatewayv2_api.api.id
  name = "${local.workspace_prefix}main"
  deployment_id = aws_apigatewayv2_deployment.main.id
}

resource "aws_api_gateway_base_path_mapping" "main" {
  api_id = aws_apigatewayv2_api.api.id
  stage_name = aws_apigatewayv2_stage.api_stage.name
  domain_name = aws_apigatewayv2_domain_name.api.domain_name
}

resource "aws_apigatewayv2_domain_name" "api" {
  domain_name = "${local.workspace_prefix}api.${data.aws_ssm_parameter.domain_name.value}"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.api_cert.arn
    endpoint_type = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_route53_record" "api" {
  name = aws_apigatewayv2_domain_name.api.domain_name
  type = "A"
  zone_id = data.aws_route53_zone.main_domain.zone_id

  alias {
    name = aws_apigatewayv2_domain_name.api.domain_name_configuration[0].target_domain_name
    zone_id = aws_apigatewayv2_domain_name.api.domain_name_configuration[0].hosted_zone_id
    evaluate_target_health = false
  }
}