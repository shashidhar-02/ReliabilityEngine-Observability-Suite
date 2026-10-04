# Contributing

## Workflow

1. Open a bug report or feature request with a reproducible example and acceptance criteria.
2. Branch from `main`: `fix/...`, `feat/...`, or `docs/...`.
3. Keep changes focused. Include behavioral tests for fixes and new application behavior.
4. Run the checks below and record results in the PR template.
5. Update deployment documentation, runbooks, and ADRs when behavior changes.
6. Open a PR; the maintainer reviews it through CODEOWNERS.

## Setup and checks

```bash
python3.11 -m venv .venv
. .venv/bin/activate
make setup PYTHON=python
make test PYTHON=python
make lint PYTHON=python
make format-check
make helm-lint
make validate
```

Go race tests require a C compiler and CGO. Terraform validation uses `init -backend=false` and does not access a cloud state backend. Container builds and scans require Docker and Trivy. GitHub Actions runs these checks on clean Linux runners.

## Standards

- Format Go with gofmt and Terraform with `terraform fmt`; lint Python with the committed flake8 configuration.
- Use bounded request and telemetry timeouts. Avoid unbounded metrics labels and global mutable test state.
- Commit dependency manifests and lock/checksum files, not downloaded libraries or executables.
- Never commit AWS credentials, state, kubeconfig, plan files, or `.env` files.
- Pin third-party Actions to reviewed commit SHAs. Let Dependabot propose upgrades.
- Every operational alert needs a runbook; significant architecture changes need an ADR.
- Use concise imperative commit messages, for example `fix: restore telemetry egress`.

## Infrastructure PRs

Describe state imports, expected cost/resource changes, deployment order, and rollback limits. Existing EKS clusters must be upgraded one minor version at a time; the configured version is the target for new installations. Public PR validation must not need AWS credentials or execute on deployment runners.

Before merge, require the Repository Quality, CodeQL, and Security Scanning checks in a branch ruleset. Actual GitHub rulesets and deployment environment permissions must be configured by a repository administrator; adding this documentation does not enable them automatically.
