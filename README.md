# NeoFS Network Monitor

Testnet network monitor. Application scrapes data from network map contract in
neo side chain. Then exposes it to prometheus instance. Then grafana visualise 
data in pretty format.

For internal usage.

## How to use 

1. (Optional) Build image of neofs-network-monitor app.

```
$ make image
```

2. (Optional) Specify neofs-net-monitor image version in `docker/docker-compose.yml`.

3. Start environment.

```
$ make up
```

To stop environment run `make down` command.

4. In grafana at `http://localhost:3000` select `NeoFS Network Monitor`
dashboard.

Supported environments:
- N3 Mainnet (`make up`)
- N3 T5 (`make up-testnet`)
- NeoFS Dev Env (`make up-devenv`)
   
## Available options

See [config examples](./config) for all available options.

You can provide a config for neofs-net-monitor:

```shell
$ neofs-net-monitor --config config.yaml
```

Also, you can provide all options using env variables.

### nep17tracker

Allows to monitor native nep17 contracts and accounts.

```yaml
nep17:
    # Native GAS contract. Possible "contract" names are Gas, GAS, gas
    - contract: "Gas" 
      label: "Gas" # Human readable contract label
      totalSupply: true # allows to return the total token supply currently available.
      balanceOf: # account list can be set via two formats in any combination
        - 3c3f4b84773ef0141576e48c3ff60e5078235891
    # Native Neo contract. Possible "contract" names are Neo, NEO, neo
    - contract: "Neo"
      label: "Neo"
      balanceOf:
        - NSPCCpw8YmgNDYWiBfXJHRfz38NDjv6WW3
        - NSPCCa2T6nc2kYcgWC2k68boyGgc9YdKsj
```

Any custom nep17 contract can be configured:

```yaml
nep17:
    # set-up contract hash
    - contract: "3c3f4b84773ef0141576e48c3ff60e5078235891"
      label: "SomeContractName"
      balanceOf:
        - NSPCCpw8YmgNDYWiBfXJHRfz38NDjv6WW3
        - NSPCCa2T6nc2kYcgWC2k68boyGgc9YdKsj
```

NeoFS balance contract can be configured with next config:

```yaml
nep17:
    - contract: "balance"
      label: "neofs_balance"
      totalSupply: true
      balanceOf:
        - NagentXDvR5c3pQ4gxXpqZjMoUpKVCUMmB
```

If contract has NNS record, you may configure tracker to monitor it. Just use NNS name in contract:

```yaml
nep17:
    - contract: "<contract_nns_name>"
      label: "<contract_name>"
      balanceOf:
        - NagentXDvR5c3pQ4gxXpqZjMoUpKVCUMmB
```

## Connect to neofs-dev-env

After `Jebudo` release monitor can be attached to 
[neofs-dev-env](https://github.com/nspcc-dev/neofs-dev-env). Go to 
`docker/docker-compose.devenv.yml` file, make sure that NeoFS contract script
hash is correct, and run `make up-devenv` command.
