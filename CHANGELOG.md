# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

### Changed
- Make flowdockAPIURL an argument to support testing
- Update README to include new argument, and test badge

### Added
- Tests, including GitHub Actions

## [0.5.0] - 2020-02-10

### Changed
- Fixed goreleaser deprecated archive to use archives
- Replaced Travis CI with GitHub Actions
- Use new Sensu SDK module

## [0.4.1] - 2019-12-17

### Changed
- Reformatted README for [Plugin Style Guide](https://github.com/sensu-plugins/community/blob/master/PLUGIN_STYLEGUIDE.md)

## [0.4.0] - 2019-08-24

### Changed
- Rewrote to use confighandler
- Updated to use modules
- Compiled with go1.12.9
- Minor documentation changes
- Remove v from version numbers for goreleaser
- Fix all references to be Flowdock (not FlowDock)

## [0.3.1] - 2019-03-25

### Changed
- Fixed issue with backend URL env variable pointing wrong config value

## [0.3.0] - 2019-03-06

### Added
- Support for annotations

### Changed
- Changed the environment variable names to be more consistent

### Added

## [0.2.0] - 2019-02-22

### Added
- added validation of backend URL
- include namespace option


## [0.1.1] - 2019-02-20

### Added
- Initial release

