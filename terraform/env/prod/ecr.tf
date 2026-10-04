resource "aws_ecr_repository" "applications" {
  for_each             = toset(["orders-api", "payments-api"])
  name                 = each.value
  image_tag_mutability = "IMMUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }
}
