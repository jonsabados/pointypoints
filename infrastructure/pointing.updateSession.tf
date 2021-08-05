module "updateSession_lambda" {
  source = "./rest-endpoint"

  aws_region    = var.aws_region
  api_id        = aws_api_gateway_rest_api.rest_pointing.id
  authorizer_id = aws_api_gateway_authorizer.authorizer.id

  name       = "updateSession"
  policy     = data.aws_iam_policy_document.session_modifying_lambda_policy.json
  lambda_env = local.session_modifying_lambda_env

  http_method = "PUT"
  resource_id = aws_api_gateway_resource.session_var.id
  full_path   = aws_api_gateway_resource.session_var.path

  request_parameters = {
    "method.request.path.session" = true
  }
}
