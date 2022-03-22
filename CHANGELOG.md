# Changelog
Changelog for NeoFS Monitor

## [Unreleased]

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

[Unreleased]: https://github.com/nspcc-dev/neofs-net-monitor/compare/v0.7.0...master
[0.7.0]: https://github.com/nspcc-dev/neofs-net-monitor/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/nspcc-dev/neofs-net-monitor/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/nspcc-dev/neofs-net-monitor/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/nspcc-dev/neofs-net-monitor/compare/v0.3.0...v0.4.0
