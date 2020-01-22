# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## Next

### Changed

- Removed panic when calling ``log.Fatal``

### 3.6.0

### Added

- Allow events that contain attributes to be decorated by the entity 
  customAttributes.

### 3.5.0

### Added

- JMX no-auth connection
- JMX constructor option to provice custom nrjmx tool executable
- JMX full connection URL

### Changed

- JMX removed line length limitation
- JMX does not support concurrency
- JMX improved test readability, removed flaky test
- JMX removed std channels unrequired close calls
- JMX better empty line comparison
- JMX removed lock
- JMX removed stderr fd closed error

### Fixed

- JMX package: fixed `nrjmx` error handling
- JMX package: fixed `nrjmx` warning handling
- JMX package support for `nrjmx` stderr multi-line responses

## 3.4.0

### Added

- Support for event attributes. Supported since infra agent version 1.5.31.

## 3.3.1

### Added

- JMX support for custom URI paths (this enables support for integrations like ForgeRock OpenDJ)

## 3.3.0

### Added

- JMX JBoss remote support for Domain-mode as default and Standalone-mode optionally.

## 3.2.0

### Added

- Adds the positive only version of `RATE` and `DELTA`, named `PRATE` and
  `PDELTA` respectively.

## 3.1.5

### Fixed

- Different [Identifier
  attributes](/blob/master/docs/entity-definition.md#identifier-attributes)
  fetch/create different entities. `integration.Entity` method fix.

## 3.1.4

### Added

- Standardise the attributes used to store the data sources: reporting-entity & reporting-endpoint under new attributes:
 `reportingEntityKey`, `reportingEndpoint`. See [doc](/docs/entity-definition.md#entity-reporting-data). 

### Changed

- New JMX option for remote URL protocol connections.
- Changed JMX open method. New `jmx.OpenWithParameters` accepts optionals parameters like SSL, remote connections, etc. Old integrations using old  methods `jmx.OpenWithSSL` and `jmx.Open` should use this new `jmx.OpenWithParameters` method instead. 

## 3.1.2

### Added

- Protocol v3: See full [documentation](docs/protocol-v3.md).


## 3.1.1

### Added

- SSL support for jmx package. Explained at https://github.com/newrelic/infra-integrations-sdk#jmx-support


## 3.1.0

### Added

- Added metadata optional decoration for entity metrics (`hostname`), [check doc](/docs/entity-definition.md)
- Added `identifier attributes` for entity metadata.


## 3.0.3

### Changed

- Replaces golint with gometalinter

### Fixed

- This release solves internal SDK dependency failures (targeting v2 via gopkg.in) as now master uses v3.
- Fixed minor lint issues


## 3.0.2

### Added

- Boolean support for *Gauge* metrics via `setMetric`
- Adds `json.Unmarsall` support to `metric.Set`
- Getters (HasMetrics, HasEvents, HasInventory) to the `args` package 
to avoid calling `All() || Metrics`
- Toolset documentation
- Tutorial using remote entities
- FAQ
- Maximum inventory key length validation
- Attributes required
- JMX concurrency support

### Fixed

- Rate & delta names collision on the Store
- Negative rate & delta values
- Prefix on the logger

### Changed

- Improved documentation
- Arguments `All` attribute to dynamic method `All()`
- NewSet does not return error
- NewSet allows Attributes to uniquely identify metric sets


## 3.0.1

### Added

- Toolset documentation
- Tutorial using remote entities
- FAQ
- Maximum inventory key length validation
- Attributes required
- JMX concurrency support

### Fixed

- Rate & delta names collision on the Store
- Negative rate & delta values
- Prefix on the logger

### Changed

- Improved documentation
- Arguments `All` attribute to dynamic method `All()`
- NewSet does not return error
- NewSet allows Attributes to uniquely identify metric sets

### Removed
- 


## 3.0.0-beta

### Added

- Protocol v2 (remote entities) support
- Concurrency support
- Configurable `Logger`
- Configurable output `Writer`
- Configurable `Storer`
- New packages `inventory`, `event` with proper constructors 
- Integration parametrized creation via optional `Option`s

### Fixed

- Nil pointer when creating a Remote entity after creating a Local entity.

### Changed

- Package `sdk` renamed to `integration`
- Package `cache` renamed to `persist` 
- Packages `metric`, `inventory`, `event` moved into `data` folder
- Replace `github.com/sirupsen/logrus` with builtin `log` package
- Update `Event` type
- Update `Integration` type

### Removed

- Protocol v1 support
- Global `log` and `cache` instances


## 2.1.0

### Added

- Adding basic travis config file

### Fixed

- Logrus package name update (to lowercase) after update


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
