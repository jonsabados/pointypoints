data "aws_iam_policy_document" "gateway_policy" {
  statement {
    sid    = "AllowLogging"
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:DescribeLogGroups",
      "logs:DescribeLogStreams",
      "logs:PutLogEvents",
      "logs:GetLogEvents",
      "logs:FilterLogEvents"
    ]
    resources = [
      "arn:aws:logs:*:*:*"
    ]
  }
}

resource "aws_iam_role" "gateway" {
  count = terraform.workspace == "default" ? 1 : 0
  name  = "${local.workspace_prefix}api_gateway_cloudwatch_global"

  assume_role_policy = data.aws_iam_policy_document.gateway_assume_role_policy.json
}

resource "aws_iam_role_policy" "gateway" {
  count = terraform.workspace == "default" ? 1 : 0
  name  = "default"
  role  = aws_iam_role.gateway[0].id

  policy = data.aws_iam_policy_document.gateway_policy.json
}

resource "aws_api_gateway_account" "pointing_gateway" {
  count               = terraform.workspace == "default" ? 1 : 0
  cloudwatch_role_arn = aws_iam_role.gateway[0].arn
}
