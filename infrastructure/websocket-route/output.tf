output "change_keys" {
  value = [
    aws_apigatewayv2_integration.lambda_integration.id,
    aws_apigatewayv2_route.route.id
  ]
}