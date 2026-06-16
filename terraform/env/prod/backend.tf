terraform {
  backend "s3" {
    bucket         = "sre-platform-terraform-state-prod"
    key            = "prod/terraform.tfstate"
    region         = "ap-south-1"
    encrypt        = true
    dynamodb_table = "sre-platform-terraform-locks-prod"
  }
}

provider "aws" {
  region = "ap-south-1"
}

module "networking" {
  source               = "../../modules/networking"
  environment          = "prod"
  vpc_cidr             = "10.0.0.0/16"
  public_subnet_count  = 3
  private_subnet_count = 3
  cluster_name         = "enterprise-sre-cluster-prod"
  tags                 = { Environment = "production" }
}

module "eks" {
  source                    = "../../modules/eks"
  cluster_name              = "enterprise-sre-cluster-prod"
  cluster_version           = "1.28"
  subnet_ids                = concat(module.networking.public_subnets, module.networking.private_subnets)
  private_subnet_ids        = module.networking.private_subnets
  cluster_security_group_id = module.networking.cluster_sg_id
  node_desired_size         = 3
  node_max_size             = 10
  node_min_size             = 3
  node_instance_types       = ["m5.large"]
}

module "iam" {
  source               = "../../modules/iam"
  oidc_provider_arn    = module.eks.oidc_provider_arn
  oidc_provider_url    = module.eks.oidc_provider_url
  namespace            = "observability"
  service_account_name = "otel-collector"
}
