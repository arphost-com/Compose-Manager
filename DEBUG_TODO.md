# Compose Manager Debugging Todo

## Confirmed Bugs

- [x] Add `web/package-lock.json` so `npm ci` and Docker web builds work.
- [ ] Prevent backup restore path traversal and cross-project restore mistakes.
- [ ] Restrict debug log reads to containers that belong to the selected project.
- [ ] Run compose config audits from the project directory so relative files and env files resolve correctly.
- [ ] Return consistent JSON API errors for authentication failures.

## Feature Work

- [ ] Show compose image/service source metadata in the API.
- [ ] Distinguish compose-built/custom services from registry image services.
- [ ] Check whether registry images are anonymously public, authenticated private, or inaccessible.
- [ ] Add private registry login support using `docker login --password-stdin`.
- [ ] Add project creation from a compose file and optional `.env` content.
- [ ] Expand the web UI to cover creation, management, updates, statistics, logging, backup, DB, security, prune, inactive mode, and bulk operations.
- [ ] Add tooltips for destructive and operational controls.

## Deployment/Validation

- [ ] Verify local Go tests and frontend build.
- [ ] Verify Docker build.
- [ ] Test on `docker02` after the GitLab project exists and deployment target is known.
