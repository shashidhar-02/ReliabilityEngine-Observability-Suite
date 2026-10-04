# Deployment and GitHub setup

## Prerequisites

Use Terraform 1.13.4, Helm 3.17, kubectl compatible with Kubernetes 1.35, Docker, and AWS CLI v2. Application development uses Go 1.26 and Python 3.11.

Bootstrap the configured S3 state buckets and DynamoDB lock tables before `terraform init`, or update the environment backend configuration to existing organization-owned resources. Enable encryption and restrict access to state. Separate staging and production state.

## GitHub configuration

Configure the `production` and `staging` GitHub environments and the `AWS_ACCOUNT_ID` secret. Configure these AWS roles with GitHub OIDC trust restricted to this repository and the matching environment subject:

| Role | Use |
| --- | --- |
| `github-actions-tf-role` | Terraform permissions and state access |
| `github-actions-ecr-role` | ECR authentication and push to the two application repositories |
| `github-actions-eks-role` | Describe cluster and approved Kubernetes deployment access |

The OIDC audience is `sts.amazonaws.com`; an example production subject is `repo:shashidhar-02/ReliabilityEngine-Observability-Suite:environment:production`. These bootstrap roles and GitHub settings are not created by this repository. Do not grant repository OIDC subjects access to an account-wide administrator role.

Register a deployment-only self-hosted Linux runner with labels `self-hosted`, `linux`, `x64`, and `eks-vpc`. It needs AWS CLI v2 and network access to the private EKS API and artifact endpoints. PR jobs run on GitHub-hosted runners, never on this runner.

Grant the deployment role Kubernetes access using the cluster's EKS access configuration or `aws-auth` mapping, including the permissions needed for namespaced application releases and the base namespace/RBAC resources. Map the explicit `sre-developers` group for developer read-only access.

Use a branch ruleset requiring PR reviews and the validation/security jobs. Configure environment reviewers, private vulnerability reporting, and code scanning through GitHub settings. Repository files alone do not turn these settings on.

## Provision infrastructure

Run the **Infrastructure Deployment** workflow on `main`. Select `staging` or `prod`. The default produces a plan; selecting `apply` applies the exact plan created in that run.

Alternatively:

```bash
make init ENV=prod
make plan ENV=prod
make apply ENV=prod
```

If ECR repositories already exist, import them before the first apply:

```bash
terraform -chdir=terraform/env/prod import 'aws_ecr_repository.applications["orders-api"]' orders-api
terraform -chdir=terraform/env/prod import 'aws_ecr_repository.applications["payments-api"]' payments-api
```

IAM role/policy names now distinguish the collector service account, preventing staging/production collisions. For existing installations, inspect the plan and migrate role references/service-account annotations as necessary. EKS version 1.35 is for new deployments; upgrade existing clusters one minor version at a time using separate reviewed plans.

Install a NetworkPolicy-capable CNI and enable its policy enforcement. Install metrics-server for the CPU HPAs. Neither a network policy manifest nor an HPA installs these dependencies.

## Deploy applications

Run **Application Deployment** on `main`. It first runs the reusable Repository Quality workflow, then publishes immutable images to the ECR repositories and deploys those exact registry/tag combinations. Re-running the same commit reuses existing immutable ECR tags.

From a VPC-connected workstation:

```bash
aws eks update-kubeconfig --region ap-south-1 --name enterprise-sre-cluster-prod
make deploy-base
make deploy-apps ECR_REGISTRY="$AWS_ACCOUNT_ID.dkr.ecr.ap-south-1.amazonaws.com" IMAGE_TAG="$(git rev-parse HEAD)"
kubectl rollout status deployment/payments-api
kubectl rollout status deployment/orders-api
```

The services are ClusterIP-only. Use port forwarding for validation. Add an ingress/gateway and a corresponding ingress NetworkPolicy before exposing Orders externally.

## Configure observability backends

The collector initially exports traces to debug logs. Replace its exporter configuration with authenticated, durable OTLP storage for a real deployment. Connect your platform Prometheus to the `/metrics` endpoints; the network policy permits scraping from the `observability` namespace. Configure alerts from [`SLOs.md`](SLOs.md) and pair each alert with a runbook. No default Slack/PagerDuty/Jira routing or error-budget deployment freeze is claimed.

The optional Terraform collector IRSA role is relevant only when using CloudWatch/X-Ray exporters. Add an annotated collector ServiceAccount and wire the role ARN before enabling those exporters; the debug exporter requires no AWS permissions.

## Rollback and release verification

Inspect `helm history <release>` and roll back to a known-good revision with `helm rollback <release> <revision> --wait`. Because Orders and Payments are separate releases, coordinate rollbacks when their API contract changes. `--atomic` rolls back each failed release independently, not both services as one transaction.

Verify health, a successful checkout, exported metrics, collector logs, replica counts, and HPA metrics after deployment. Do not label a release production-verified until this live verification has succeeded in your account.
