resource "aws_dynamodb_table" "session_store" {
  hash_key = "SessionId"
  name = "${local.workspace_prefix}Session"
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = "SessionId"
    type = "S"
  }

  ttl {
    enabled = "true"
    attribute_name = "Expiration"
  }

  tags = {
    Workspace = terraform.workspace
  }
}