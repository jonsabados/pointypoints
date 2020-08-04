resource "aws_dynamodb_table" "session_store" {
  hash_key = "SessionID"
  name = "${local.workspace_prefix}Session"
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = "SessionID"
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