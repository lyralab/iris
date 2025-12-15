# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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