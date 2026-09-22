# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2026-09-22
### Fixed
- SMS.ir API key configuration key name, so the token loads from `apikey`
- Kavenegar API token configuration key name, so the token loads from `token`

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