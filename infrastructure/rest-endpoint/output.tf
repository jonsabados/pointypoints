output "change_keys" {
  value = [
    var.resource_id,
    aws_api_gateway_method.method.id,
    aws_api_gateway_integration.integration.id,
    var.authorizer_id,
  ]
}