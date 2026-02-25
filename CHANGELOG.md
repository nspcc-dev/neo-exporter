# Changelog
Changelog for NeoFS Monitor

## [Unreleased]

### Added

### Changed

### Removed

### Fixed

## [0.15.2] - 2026-02-25

### Changed
- Go 1.25+ is required to build now (#172)
- Updated golang.org/x/crypto dependency from 0.42.0 to 0.45.0 (#184)
- Updated NeoFS SDK dependency to RC17 (#188)
- Updated github.com/nspcc-dev/neofs-contract dependency to v0.26.1 (#188)
- Updated NeoGo dependency to 0.117.0 (#188)
- Updated github.com/nspcc-dev/locode-db dependency to v0.8.2 (#188)
- Updated go.uber.org/zap dependency to v1.27.1 (#188)
- Updated golang.org/x/term to v0.40.0 (#188)

### Fixed
- Return code on graceful shutdown of uninited process (#189)

## [0.15.1] - 2025-11-11

### Fixed
- Failure to start if any of specified endpoints is not responding (#181)

## [0.15.0] - 2025-10-15

### Added
- Node capacity and overall cluster size metric (#177)
- Total and per-container object counts and sizes (#176)

### Changed
- github.com/nspcc-dev/neofs-sdk-go v1.0.0-rc.14 => v1.0.0-rc.15.0.20251014091731-abf7dc3b4b64 (#177)
- github.com/nspcc-dev/neofs-contract v0.23.0 => v0.24.0

## [0.14.1] - 2025-09-23

### Changed
- Go 1.24+ is required to build now (#160)
- github.com/nspcc-dev/neofs-sdk-go v1.0.0-rc.13 => v1.0.0-rc.14 (#160)
- github.com/nspcc-dev/locode-db v0.7.0 => v0.8.1 (#160)
- github.com/nspcc-dev/hrw/v2 v2.0.3 => v2.0.4 (#160)
- github.com/nspcc-dev/neofs-contract v0.21.0 => v0.23.0 (#160)
- github.com/prometheus/client_golang v1.20.2 => v1.23.2 (#160)
- github.com/spf13/viper v1.19.0 => v1.21.0 (#160)
- golang.org/x/term v0.34.0 => v0.35.0 (#160)
- github.com/multiformats/go-multiaddr v0.13.0 => v0.16.1 (#160)
- Don't stop app starting and wait until pool connected (#171)
- Updated NeoGo dependency to 0.113.0 (#175)

## [0.14.0] - 2025-04-14

### Removed
- V1 netmap logic (#166)

## [0.13.0] - 2025-03-25

### Added
- Netmap v2 candidate metric (#159)

### Changed
- Log sampling is disabled now (#156)
- Go 1.23+ is required to build now (#148)
- Updated github.com/nspcc-dev/neofs-sdk-go v1.0.0-rc.12 => v1.0.0-rc.13 (#159)
- Updated github.com/nspcc-dev/neofs-contract v0.20.0 => v0.21.0 (#159)

### Removed
- github.com/nspcc-dev/neofs-api-go dependency (#159)

Please notice that "side_chain_supply" metric was renamed to "fs_chain_supply"
in this release.

## [0.12.1] - 2024-09-03

### Changed
- Go 1.22+ is required to build now (#149, #150)
- Dropped 'v' prefix from version metric (#151)
- Updated github.com/nspcc-dev/locode-db dependency to 0.7.0 (#150)
- Updated github.com/nspcc-dev/hrw/v2 dependency to 2.0.3 (#150)
- Updated github.com/prometheus/client_golang dependency to v1.20.2 (#150)
- Default dial timeout to one minute (#152)
- Default metric scraping interval to 15s (#152)

## [0.12.0] - 2024-06-19

Please notice that "neofs_net_monitor" prefix was changed to "neo_exporter"
and update accordingly.

### Added
- Version metric (#143)

### Fixed
- Panics on RPC reconnection failure (#141)

### Changed
- "neofs_net_monitor" prefix to "neo_exporter" (#143)
- Timestamps are no longer produced in logs if not running with TTY (#142)

## [0.11.3] - 2024-06-18

### Fixed
- Panic on RPC reconnection failure (#137)

## [0.11.2] - 2024-06-13

This release adds support for the Domovoi hardfork with no other functional
changes.

### Changed
- github.com/nspcc-dev/neo-go dependency from 0.106.0 to 0.106.2 (#135)
- NeoFS SDK dependency to RC12 (#135)

## [0.11.1] - 2024-05-22

### Added
- arm64 and darwin builds (#132)

### Changed
- google.golang.org/protobuf dependency from 1.32.0 to 1.33.0 (#129)
- golang.org/x/net dependency from 0.21.0 to 0.23.0 (#131)
- github.com/nspcc-dev/neo-go dependency from 0.105.1 to 0.106.0 (#133)

## [0.11.0] - 2024-03-07

Please notice that monitored NEP-17 accounts are exported in address form now
and update your settings.

### Fixed
- Notary balance tracking (#122)

### Changed
- Go 1.20+ is required to build now (#114)
- Accounts are exported as address now (#123)

### Removed
- "label" label for tracked NEP-17 contracts (#126)

## [0.10.0] - 2024-02-16

### Added
- Height and state data export for a set of configured nodes (#96)
- NEP-17 balance tracking (#105)

### Changed
- Go 1.19+ is required to build now (#92)
- Updated all dependencies (#92, #109)
- Contract wrappers are used now to interact with blockchain (#98)
- Usage of Locode DB Go package (#100)
- Configuration supports only one chain in a moment (#103)
- The tool is known as neo-exporter now (#112)

### Removed
- Notary-less mode for FS chains (#97)
- Locode DB configuration options (#100)

### Upgrading from v0.9.5

The configuration sections `mainnet` and  `morph` were replaced with similar `chain` sections. To choice between
main (Neo) chain and side (NeoFS) chain, use `chain.fschain` option. If true, exporter connects to the NeoFS chain,
otherwise, to the Neo chain. It no longer watches two chains at once, so to monitor NeoFS you need two instances
of the tool.

## [0.9.5] - 2022-12-29

### Changed
- Removed neofs-node dependency (#75)
- Fixed compatibility with the netmap contract version 0.16.0 (#81)

## [0.9.0] - 2022-07-27

### Changed
- Direct contract communication instead of having neofs-node wrappers (#73)
- Using `localhost` in docs (#76)
- T5 Network support replaced T4 (#78)

## [0.8.0] - 2022-05-24

### Added
- `neofs_net_monitor_containers_number` metric (#71)

## [0.7.1] - 2022-03-30

### Added
- Support rpc connection pool (#63)

## [0.7.0] - 2022-03-22

### Fixed
- Use `$datasource` in grafana (#53)

### Added
- Log of storage node appearance in network map (#26)
- Dropdown list to pick storage node public keys (#4)
- Alphabet divergence metrics (#60)

### Changed
- Update grafana image to v8.4.2 (#59)
- Update LOCODE db to v0.2.1

## [0.6.0] - 2021-11-02

### Fixed
- Support updated NNS contract (#49)

### Added
- Description for `make locode` command (#44)
- Asset supply metrics of main and side chain economy (#45)

### Upgrading from v0.5.0
Specify main chain NeoFS contract address in `NEOFS_NET_MONITOR_CONTRACTS_NEOFS`
to enable main chain supply metric.

## [0.5.0] - 2021-10-08

### Added
- Geo position is retrieved from NeoFS locode database (#17)
- Prometheus labels that distinguish different networks (#31)
- `VERSION` file for builds without `git` (#37)
- Contract script-hashes auto negotiation (#33)
- Reusing `neo-go` client (#3)
- Notary balances of SN and IR (#36)

### Fixed
- Failing on initial startup (#32)

### Removed
- Fetching geo position from external sources (#17)

### Upgrading from v0.4.0
All `NEOFS_NET_MONITOR_CONTRACTS_*` envs now are optional if corresponding chain's NNS
contract contains corresponding contract script hashes.
External geoIP service support is dropped and `NEOFS_NET_MONITOR_GEOIP_*` envs are not
used anymore.
`NEOFS_NET_MONITOR_LOCODE_DB_PATH` env has been added. It is path to NeoFS locode
[database](https://github.com/nspcc-dev/neofs-locode-db). Optional.

## [0.4.0] - 2021-07-28

### Added
- Alphabet node gas balances in main chain (#8);
- Netmap new/dropped node candidates (#12).

### Fixed
- Missing TLS certificates in docker image (#14). 

### Removed
- Explicit `Prometheus` dependency in grafana (#21).

### Updated
- Hosts and script-hashes in docker files for RC4 (#23);
- Go to 1.16 (#16).

### Upgrading from v0.3.0
There are two NEO RPC nodes that are used in `neofs-net-monitor`(side and main chain) now.
`NEOFS_NET_MONITOR_RPC_ENDPOINT` and `NEOFS_NET_MONITOR_RPC_DIAL_TIMEOUT` envs were deleted.
Use `NEOFS_NET_MONITOR_MAINNET_RPC_ENDPOINT`, `NEOFS_NET_MONITOR_MAINNET_RPC_DIAL_TIMEOUT`
to establish connection to the RC4 main chain and `NEOFS_NET_MONITOR_MORPH_RPC_ENDPOINT`,
`NEOFS_NET_MONITOR_MORPH_RPC_DIAL_TIMEOUT` for the side chain instead.
`NEOFS_NET_MONITOR_CONTRACTS_PROXY` env is now optional for notary disabled environments.

[Unreleased]: https://github.com/nspcc-dev/neo-exporter/compare/v0.15.2...master
[0.15.2]: https://github.com/nspcc-dev/neo-exporter/compare/v0.15.1...v0.15.2
[0.15.1]: https://github.com/nspcc-dev/neo-exporter/compare/v0.15.0...v0.15.1
[0.15.0]: https://github.com/nspcc-dev/neo-exporter/compare/v0.14.1...v0.15.0
[0.14.1]: https://github.com/nspcc-dev/neo-exporter/compare/v0.14.0...v0.14.1
[0.14.0]: https://github.com/nspcc-dev/neo-exporter/compare/v0.13.0...v0.14.0
[0.13.0]: https://github.com/nspcc-dev/neo-exporter/compare/v0.12.1...v0.13.0
[0.12.1]: https://github.com/nspcc-dev/neo-exporter/compare/v0.12.0...v0.12.1
[0.12.0]: https://github.com/nspcc-dev/neo-exporter/compare/v0.11.3...v0.12.0
[0.11.3]: https://github.com/nspcc-dev/neo-exporter/compare/v0.11.2...v0.11.3
[0.11.2]: https://github.com/nspcc-dev/neo-exporter/compare/v0.11.1...v0.11.2
[0.11.1]: https://github.com/nspcc-dev/neo-exporter/compare/v0.11.0...v0.11.1
[0.11.0]: https://github.com/nspcc-dev/neo-exporter/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/nspcc-dev/neo-exporter/compare/v0.9.5...v0.10.0
[0.9.5]: https://github.com/nspcc-dev/neo-exporter/compare/v0.9.0...v0.9.5
[0.9.0]: https://github.com/nspcc-dev/neo-exporter/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/nspcc-dev/neo-exporter/compare/v0.7.1...v0.8.0
[0.7.1]: https://github.com/nspcc-dev/neo-exporter/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/nspcc-dev/neo-exporter/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/nspcc-dev/neo-exporter/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/nspcc-dev/neo-exporter/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/nspcc-dev/neo-exporter/compare/v0.3.0...v0.4.0
