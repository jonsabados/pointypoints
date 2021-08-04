locals {
  session_socket_index_name = "${local.workspace_prefix}SessionSockets"
}

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

  attribute {
    name = "SocketID"
    type = "S"
  }

  global_secondary_index {
    name            = local.session_socket_index_name
    hash_key        = "SocketID"
    projection_type = "KEYS_ONLY"
  }

  ttl {
    enabled        = "true"
    attribute_name = "Expiration"
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_dynamodb_table" "profile_store" {
  hash_key     = "UserID"
  name         = "${local.workspace_prefix}Profile"
  billing_mode = "PAY_PER_REQUEST"

  attribute {
    name = "UserID"
    type = "S"
  }

  tags = {
    Workspace = terraform.workspace
  }
}