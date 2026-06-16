# Dummy root module to force VS Code Terraform Extension to index sub-modules

module "eks_index" {
  source = "./terraform/modules/eks"
  
  cluster_name              = "dummy"
  subnet_ids                = []
  private_subnet_ids        = []
  cluster_security_group_id = "dummy"
}

module "networking_index" {
  source = "./terraform/modules/networking"
  
  environment = "dummy"
  vpc_cidr    = "10.0.0.0/16"
  cluster_name = "dummy"
}
