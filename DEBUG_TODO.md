# Compose Manager Debugging Todo

## Confirmed Bugs

- [x] Add `web/package-lock.json` so `npm ci` and Docker web builds work.
- [x] Prevent backup restore path traversal and cross-project restore mistakes.
- [x] Restrict debug log reads to containers that belong to the selected project.
- [x] Run compose config audits from the project directory so relative files and env files resolve correctly.
- [x] Return consistent JSON API errors for authentication failures.

## Feature Work

- [x] Show compose image/service source metadata in the API.
- [x] Distinguish compose-built/custom services from registry image services.
- [x] Check whether registry images are anonymously public, authenticated private, or inaccessible.
- [x] Add private registry login support using `docker login --password-stdin`.
- [x] Add project creation from a compose file and optional `.env` content.
- [x] Expand the web UI to cover creation, management, updates, statistics, logging, backup, DB, security, prune, inactive mode, and bulk operations.
- [x] Add tooltips for destructive and operational controls.

## Deployment/Validation

- [x] Verify local Go tests and frontend build.
- [ ] Verify Docker build.
- [x] Add GitLab pipeline for docker02 validation, security scanning, deployment, and smoke checks.
- [ ] Test on `docker02` after the GitLab project exists and deployment target is known.
