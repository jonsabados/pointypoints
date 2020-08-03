resource "aws_dynamodb_table" "global_locks" {
  hash_key = "LockID"
  name = "${local.workspace_prefix}Locks"
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = "LockID"
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