# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.3] - 2024-04-26

## Added

- Snailrace Core
- Perishable Snail Trading Cards

## [0.2.2] - 2024-04-22

## Fixed

- Logging consistency to allow logs to be managed in Grafana Loki

## [0.2.1] - 2024-04-22

### Fixed

- Blackjack would get stuck out once the session is finished. Handling dealer's state better.
- Reduced the number of times the Dealer's state gets set to not upset Discords API.

## [0.2.0] - 2024-04-17

### Added

- Wallet System (#16)
- Persistent Database (#14)
- Blackjack (#18)
- Trading Cards v1 (#20)
- Testing Workflow (#22)
- Blackjack Achievements (#23)
- Notify On Startup (#17)

### Changed

- Prod database has moved from sqlite3 to postgreSQL
- Test database uses a CGO free sqlite3 driver for running unit tests

## [0.1.1] - 2024-04-10

### Added

- Use [gorm](https://gorm.io/) for handling database interactions
- Runtime checks to force application function implementations

### Changed

- Applications and Application Rules are now all Applications
- Applications can choose what kind of application they are through the `GetType()` so they dont have to implement all the different callback funcitons
- The `framework.Context` has been split into multiple types for each application callback so developers can't call things that dont have any values.

### Fixed

- Deploy stops and removes the previous running container before trying to deploy
- Removed dead code
- For message type applications, the right message author is being pulled
- Autopin requires 5 pins instead of just 1
- Deploy requires the build stage to finish before it can deploy

## [0.1.0] - 2024-04-05
 
### Added

- Applications to build commands
- Application Rules for Moderation usecases
- Database Connectivity
- Base Framework for interactiving with the Discord API
- A basic deployment system to run the Bot from tagging

[unreleased]: https://github.com/aussiebroadwan/tony/compare/v0.2.3...HEAD
[0.2.3]: https://github.com/aussiebroadwan/tony/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/aussiebroadwan/tony/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/aussiebroadwan/tony/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/aussiebroadwan/tony/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/aussiebroadwan/tony/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/aussiebroadwan/tony/releases/tag/v0.1.0
