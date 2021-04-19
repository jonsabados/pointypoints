resource "aws_api_gateway_resource" "vote_resource" {
  rest_api_id = aws_api_gateway_rest_api.rest_pointing.id
  parent_id   = aws_api_gateway_resource.user_var.id
  path_part   = "vote"
}

module "vote_lambda" {
  source = "./rest-endpoint"

  aws_region = var.aws_region
  api_id     = aws_api_gateway_rest_api.rest_pointing.id

  name       = "vote"
  policy     = data.aws_iam_policy_document.session_modifying_lambda_policy.json
  lambda_env = local.session_modifying_lambda_env

  http_method = "PUT"
  resource_id = aws_api_gateway_resource.vote_resource.id
  full_path   = aws_api_gateway_resource.vote_resource.path

  request_parameters = {
    "method.request.path.session"    = true
    "method.request.path.user" = true
  }
}
