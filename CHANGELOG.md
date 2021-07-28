# Changelog
Changelog for NeoFS Monitor

## [Unreleased]

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

[Unreleased]: https://github.com/nspcc-dev/neofs-net-monitor/compare/v0.4.0...master
[0.4.0]: https://github.com/nspcc-dev/neofs-net-monitor/compare/v0.3.0...v0.4.0
