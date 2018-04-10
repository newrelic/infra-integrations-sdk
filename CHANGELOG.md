# Change Log
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

### Added
- Protocol v2 support
- Loggers can now be created from the log package.
- Already created loggers can be configured from the log package.
- Updated github.com/sirupsen/logrus dependency to v1.*
- Add methods for creating events
### Changed
- **Breaking change**: Update `Event` type
- **Breaking change**: Update `Integration` struct type
### Fixed
- github.com/Sirupsen/logrus is now github.com/sirupsen/logrus since v1.0.0.

## 2.0.0 (2017-10-11)
### Fixed
- Allow executing JMX queries to a server without authentication required
### Changed
- **Breaking change**: Update `jmx.Query` function adding timeout argument
- Improve documentation

## 1.0.0 (2017-07-20)

This is the same version as 0.4.1, we consider this package "stable" and we can
release v1.0.0

## 0.4.1 (2017-07-19)
### Fixed
- Use absolute path calling nrjmx binary

### Changed
- Update tutorial with changes about metric names, event types and some other improvements

## 0.4.0 (2017-07-13)
### Added
- Add a JSON type for Arguments

### Changed
- Clear internal objects after in `integration.Publish()` method
- Remove provider argument in `NewMetricSet` function

## 0.3.1 (2017-06-22)
### Fixed
- Allow to set multiple values for the same key with `Inventory.SetItem()` function
- Variadic arguments to log methods
- Increase buffer length for reading from JMX

## 0.3.0 (2017-06-19)
### Added
- New license file
- New `SetItem` function for inventory
- New `Inventory` struct type

### Changed
- Rename `COUNTER` metric type to `RATE`
- Improve jmx package tests
- Rename `AddMetric` to `SetMetric`
- Improve the interoperability with `nrjmx` tool
- Update some documentation strings

## 0.2
### Fixed
- Fix sampling cache path on windows

## 0.1
### Added
- Initial release
