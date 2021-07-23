resource "aws_api_gateway_resource" "session_votes_resource" {
  rest_api_id = aws_api_gateway_rest_api.rest_pointing.id
  parent_id   = aws_api_gateway_resource.session_var.id
  path_part   = "votes"
}

module "clearVotes_lambda" {
  source = "./rest-endpoint"

  aws_region = var.aws_region
  api_id     = aws_api_gateway_rest_api.rest_pointing.id

  name       = "clearVotes"
  policy     = data.aws_iam_policy_document.session_modifying_lambda_policy.json
  lambda_env = local.session_modifying_lambda_env

  http_method = "DELETE"
  resource_id = aws_api_gateway_resource.session_votes_resource.id
  full_path   = aws_api_gateway_resource.session_votes_resource.path

  request_parameters = {
    "method.request.path.session" = true
  }
}
