data "aws_ssm_parameter" "ui_bucket_name" {
  name = "pointypoints.uibucket"
}

data "aws_ssm_parameter" "google_site_verification_record" {
  name = "pointypoints.google.verification"
}