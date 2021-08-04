resource "aws_api_gateway_resource" "profile" {
  rest_api_id = aws_api_gateway_rest_api.rest_pointing.id
  parent_id   = aws_api_gateway_rest_api.rest_pointing.root_resource_id
  path_part   = "profile"
}