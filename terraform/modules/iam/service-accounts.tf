variable "oidc_provider_arn" {
  description = "OIDC Provider ARN for EKS"
  type        = string
}

variable "oidc_provider_url" {
  description = "OIDC Provider URL for EKS"
  type        = string
}

variable "namespace" {
  description = "Kubernetes Namespace for the service accounts"
  type        = string
  default     = "default"
}

variable "service_account_name" {
  description = "Name of the Kubernetes service account"
  type        = string
  default     = "otel-collector-sa"
}

# Example IAM Role for OpenTelemetry Collector to write to CloudWatch/X-Ray (if needed)
resource "aws_iam_role" "otel_collector" {
  name = "eks-otel-collector-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Federated = var.oidc_provider_arn
      }
      Action = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:${var.namespace}:${var.service_account_name}"
        }
      }
    }]
  })
}

resource "aws_iam_policy" "otel_collector_policy" {
  name        = "eks-otel-collector-policy"
  description = "Policy for OpenTelemetry Collector"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "logs:PutLogEvents",
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:DescribeLogStreams",
        "logs:DescribeLogGroups",
        "xray:PutTraceSegments",
        "xray:PutTelemetryRecords",
        "xray:GetSamplingRules",
        "xray:GetSamplingTargets",
        "xray:GetSamplingStatisticSummaries"
      ]
      Resource = "*"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "otel_collector_attach" {
  role       = aws_iam_role.otel_collector.name
  policy_arn = aws_iam_policy.otel_collector_policy.arn
}
