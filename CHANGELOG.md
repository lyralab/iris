# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2026-10-14
### Added
- Dark theme with a header toggle; follows the OS setting until a theme is picked

### Changed
- Web lint also runs stylelint and rejects hard-coded hex colours outside the shared theme styles

### Removed
- Unused `App.css`

## [1.1.0] - 2026-10-14
### Added
- Alerts page and dashboard auto-refresh on a saved interval
- Latest notifications per provider (`GET /v0/providers/:name/messages`) and a provider messages view
- Lint web sources in merge request pipelines

### Changed
- Alerts are sorted newest first on the dashboard and alerts page, with the same fallbacks as the API
- Receptor cache is refreshed after group changes and after a user's contact details change
- A failed SMS provider falls back to the next active SMS provider by priority, then name

### Fixed
- Duplicate alert delivery and missing Telegram delivery status
- Overlapping alert refreshes and stale responses overwriting newer alerts in the web UI
- Receptors leaking between providers, so Telegram no longer receives SMS numbers
- Asiatech panic on a failed request; a host without a scheme is rejected with a clear error
- Selecting a provider that is active in the database but disabled in config no longer panics

## [1.0.2] - 2026-09-22
### Fixed
- Drop the unused Mattermost server module so image builds no longer download `server/v8`

## [1.0.1] - 2026-09-22
### Fixed
- SMS.ir API key configuration key name, so the token loads from `apikey`
- Kavenegar API token configuration key name, so the token loads from `token`
- Postgres password configuration key name, so the password loads from `password`
- Mattermost message attachments for the updated SDK

### Changed
- Upgrade Go module dependencies
- Pin vulnerable web dependencies

## [1.0.0] - 2026-09-22
### Added
- Remove a user from a group from the API and the group management page
- Confirmation dialog before destructive actions in user and group management
- Group descriptions on group list responses
- Telegram, Mattermost, and Bale ids on user list and user detail responses
- Bale provider settings in the environment example

### Changed
- Telegram and Bale ids accept negative channel ids
- Clearing a contact field on a user now saves as empty
- Shared dashboard styles across login, alerts, users, groups, providers, and profile
- HTTP server listens on all interfaces

### Fixed
- Group membership removal used the wrong columns and a soft delete, which blocked adding the user again

## [0.1.4] - 2026-09-21
### Added
- Implement Bale provider for notifications
### Changed
- Upgrade Go from 1.24 to 1.27
- Switch Docker build to Chainguard images (go/node/static)

## [0.1.3] - 2026-02-21
### Changed
- Fix gitlab ci bug

## [0.1.2] - 2026-02-21
### Changed
- Resolve bug in koanf configuration loading

## [0.1.1] - 2026-02-21
### Changed
- Fix typo

## [0.1.0] - 2026-02-15
### Added
- Create deployment files and ready for first release

## [0.0.11] - 2026-02-14
### Changed
- Change http to server in configuration file

## [0.0.10] - 2026-02-14
### Changed
- Fix bux in startup provider configuration

## [0.0.9] - 2026-02-13
### Added
- Support environment variables 

## [0.0.8] - 2026-02-01
### Changed
- Implement Asiatech provider for SMS notifications

## [0.0.7] - 2025-12-31
### Changed
- Fix missing migration file

## [0.0.6] - 2025-12-15
### Changed
- Fix text message formatting for telegram provider and sms providers

## [0.0.5] - 2025-12-15
### Changed
- Update docker base image to latest stable version(golang:1.25-alpine)

## [0.0.4] - 2025-12-15
### Changed
- Support for multiple method for each alert
- Fix bug in alert delivery system (send to user multiple times)

## [0.0.3] - 2025-12-10
### Added
- Add mail as a notification channel
- Add mattermost as a notification channel

### Changed
- Refactor message for each provider for better readability

## [0.0.2] - 2025-12-08 
### Added
- Add telegram as a notification channel
### Changed
- Fixed notification delivery issues
- Improved profile page web dashboard

## [0.0.1] - 2025-11-04
### Added
- Initial release
- Alert management system
- Multi-channel notifications
- User and role management
- Web dashboard
- AlertManager integration