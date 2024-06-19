# Changelog
Changelog for NeoFS Monitor

## [Unreleased]

Please notice that "neofs_net_monitor" prefix was changed to "neo_exporter"
and update accordingly.

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

[Unreleased]: https://github.com/nspcc-dev/neo-exporter/compare/v0.11.3...master
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
