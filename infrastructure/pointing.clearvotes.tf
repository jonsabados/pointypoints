module "clearVotes_lambda" {
  source = "./websocket-route"

  aws_region = var.aws_region

  api_id = aws_apigatewayv2_api.websockets_pointing.id
  name   = "clearVotes"
  route  = "clearVotes"

  policy = data.aws_iam_policy_document.session_modifying_lambda_policy.json

  lambda_env = local.session_modifying_lambda_env
}
