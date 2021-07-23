resource "aws_s3_bucket" "ui_bucket" {
  bucket = "${local.workspace_prefix}${data.aws_ssm_parameter.ui_bucket_name.value}"
  acl    = "public-read"

  website {
    index_document = "index.html"
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_acm_certificate" "ui_cert" {
  domain_name = "${local.workspace_domain_prefix}${data.aws_ssm_parameter.domain_name.value}"
  subject_alternative_names = [
    "${terraform.workspace == "default" ? "www." : "www-"}${local.workspace_domain_prefix}${data.aws_ssm_parameter.domain_name.value}"
  ]
  validation_method = "DNS"

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_route53_record" "ui_cert_cert_verification_record" {
  count   = 2
  name    = aws_acm_certificate.ui_cert.domain_validation_options[count.index].resource_record_name
  type    = aws_acm_certificate.ui_cert.domain_validation_options[count.index].resource_record_type
  zone_id = data.aws_route53_zone.main_domain.id
  records = [
  aws_acm_certificate.ui_cert.domain_validation_options[count.index].resource_record_value]
  ttl = 300
}

resource "aws_cloudfront_origin_access_identity" "default" {}

resource "aws_cloudfront_distribution" "ui_cdn" {
  enabled             = true
  wait_for_deployment = false
  price_class         = "PriceClass_100"
  default_root_object = "index.html"
  aliases = [
    "${local.workspace_domain_prefix}${data.aws_ssm_parameter.domain_name.value}",
    "${terraform.workspace == "default" ? "www." : "www-"}${local.workspace_domain_prefix}${data.aws_ssm_parameter.domain_name.value}"
  ]

  default_cache_behavior {
    allowed_methods = [
      "HEAD",
      "GET"
    ]
    cached_methods = [
      "HEAD",
      "GET"
    ]
    target_origin_id       = "ui_bucket"
    viewer_protocol_policy = "redirect-to-https"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }
  }

  custom_error_response {
    error_code         = 404
    response_code      = 200
    response_page_path = "/index.html"
  }

  origin {
    origin_id   = "ui_bucket"
    domain_name = aws_s3_bucket.ui_bucket.bucket_regional_domain_name

    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.default.cloudfront_access_identity_path
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    ssl_support_method  = "sni-only"
    acm_certificate_arn = aws_acm_certificate.ui_cert.arn
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_route53_record" "default_domain_name" {
  name    = "${local.workspace_domain_prefix}${data.aws_ssm_parameter.domain_name.value}"
  type    = "A"
  zone_id = data.aws_route53_zone.main_domain.zone_id

  alias {
    name                   = aws_cloudfront_distribution.ui_cdn.domain_name
    zone_id                = aws_cloudfront_distribution.ui_cdn.hosted_zone_id
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "www_domain_name" {
  name    = terraform.workspace == "default" ? "www" : "www-${terraform.workspace}"
  type    = "CNAME"
  zone_id = data.aws_route53_zone.main_domain.zone_id
  records = [aws_cloudfront_distribution.ui_cdn.domain_name]
  ttl     = 60
}

resource "aws_route53_record" "domain_txt_records" {
  count   = terraform.workspace == "default" ? 1 : 0
  name    = data.aws_ssm_parameter.domain_name.value
  type    = "TXT"
  zone_id = data.aws_route53_zone.main_domain.id
  records = [data.aws_ssm_parameter.google_site_verification_record.value]
  ttl     = 900
}