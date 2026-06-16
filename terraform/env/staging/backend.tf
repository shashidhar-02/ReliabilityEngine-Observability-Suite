terraform {
  backend "s3" {
    bucket         = "sre-platform-terraform-state-staging"
    key            = "staging/terraform.tfstate"
    region         = "ap-south-1"
    encrypt        = true
    dynamodb_table = "sre-platform-terraform-locks-staging"
  }
}

provider "aws" {
  region = "ap-south-1"
}

module "networking" {
  source               = "../../modules/networking"
  environment          = "staging"
  vpc_cidr             = "10.1.0.0/16"
  public_subnet_count  = 2
  private_subnet_count = 2
  cluster_name         = "enterprise-sre-cluster-staging"
  tags                 = { Environment = "staging" }
}

module "eks" {
  source                    = "../../modules/eks"
  cluster_name              = "enterprise-sre-cluster-staging"
  cluster_version           = "1.28"
  subnet_ids                = concat(module.networking.public_subnets, module.networking.private_subnets)
  private_subnet_ids        = module.networking.private_subnets
  cluster_security_group_id = module.networking.cluster_sg_id
  node_desired_size         = 2
  node_max_size             = 3
  node_min_size             = 1
  node_instance_types       = ["t3.medium"]
}

module "iam" {
  source               = "../../modules/iam"
  oidc_provider_arn    = module.eks.oidc_provider_arn
  oidc_provider_url    = module.eks.oidc_provider_url
  namespace            = "observability"
  service_account_name = "otel-collector-staging"
}
