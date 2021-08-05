module "newSession_lambda" {
  source = "./rest-endpoint"

  aws_region    = var.aws_region
  api_id        = aws_api_gateway_rest_api.rest_pointing.id
  authorizer_id = aws_api_gateway_authorizer.authorizer.id

  name       = "newSession"
  policy     = data.aws_iam_policy_document.session_modifying_lambda_policy.json
  lambda_env = local.session_modifying_lambda_env

  http_method = "POST"
  resource_id = aws_api_gateway_resource.session_path.id
  full_path   = aws_api_gateway_resource.session_path.path

  request_parameters = {}
}
