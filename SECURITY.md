# Security policy

The default branch receives security fixes. Use immutable application image tags and supported runtime and Kubernetes versions.

## Report a vulnerability

Use GitHub private vulnerability reporting when enabled:

https://github.com/shashidhar-02/ReliabilityEngine-Observability-Suite/security/advisories/new

If that option is unavailable, contact the maintainer privately through a contact channel listed on their GitHub profile. Do not post exploit details or credentials in a public issue.

Include the affected commit/component, impact, reproduction steps, and suggested mitigation. A response timeline is not guaranteed; coordinate public disclosure with the maintainer.

## Controls

CI uses read-only permissions except explicitly scoped OIDC and security-reporting jobs. Deployment runners are reserved for reviewed main-branch workflows. Secrets remain in GitHub environments or an approved secret manager. HIGH/CRITICAL image vulnerabilities fail scanning jobs; a passing scan is not a guarantee of complete security.
