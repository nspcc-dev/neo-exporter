# Neo Exporter

Export various data from Neo chains as Prometheus metrics. Special support is
provided for NeoFS contracts (deployed into FS chains). Metrics are updated
regularly and can then be scraped/stored/analysed/visualized in various ways
(VictoriaMetrics/Grafana).

## Available options

See [config examples](./config) for all available options.

You can provide a config for neo-exporter:

```shell
$ neo-exporter --config config.yaml
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
