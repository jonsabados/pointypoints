resource "aws_dynamodb_table" "session_store" {
  hash_key     = "SessionID"
  range_key    = "RangeKey"
  name         = "${local.workspace_prefix}Session"
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = "SessionID"
    type = "S"
  }

  attribute {
    name = "RangeKey"
    type = "S"
  }

  ttl {
    enabled        = "true"
    attribute_name = "Expiration"
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_dynamodb_table" "session_interest_store" {
  hash_key     = "ConnectionID"
  name         = "${local.workspace_prefix}SessionInterest"
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = "ConnectionID"
    type = "S"
  }

  ttl {
    enabled        = "true"
    attribute_name = "Expiration"
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_dynamodb_table" "session_watcher_store" {
  hash_key     = "SessionID"
  name         = "${local.workspace_prefix}SessionWatcher"
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = "SessionID"
    type = "S"
  }

  ttl {
    enabled        = "true"
    attribute_name = "Expiration"
  }

  tags = {
    Workspace = terraform.workspace
  }
}