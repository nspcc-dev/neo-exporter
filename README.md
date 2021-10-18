# NeoFS Network Monitor

Testnet network monitor. Application scrapes data from network map contract in
neo side chain. Then exposes it to prometheus instance. Then grafana visualise 
data in pretty format.

For internal usage.

## How to use 

1. Build image of neofs-network-monitor app.

```
$ make image
...
Successfully built 22b63620bc9d
Successfully tagged nspccdev/neofs-net-monitor:0.4.0
```

2. (Optional) Specify neofs-net-monitor image version in `docker/docker-compose.yml`.

3. Start environment.

```
$ make up
Creating network "docker_monitor-net" with driver "bridge"
Creating prometheus        ... done
Creating grafana           ... done
Creating neofs-net-monitor ... done
```

To stop environment run `make down` command.

4. In grafana at `http://127.0.0.1:3000` select `NeoFS Network Monitor`
dashboard.
   
## Available options

```
// WIF/hex/path to binary private key for the NEO client. If not set, it is randomly generated at startup.
NEOFS_NET_MONITOR_KEY=KyH4ASQ1tmm7q9eQKiSzCSH6kxNVbUe3B41EeLaJ15UoMwgZw3Zk 

// Sidechain NEO RPC related configuration.
NEOFS_NET_MONITOR_MORPH_RPC_ENDPOINT=https://rpc1-morph.preview4.nspcc.ru:24333
NEOFS_NET_MONITOR_MORPH_RPC_DIAL_TIMEOUT=5s

// Mainchain NEO RPC related configuration.
NEOFS_NET_MONITOR_MAINNET_RPC_ENDPOINT=https://rpc1-main.preview4.nspcc.ru:24333
NEOFS_NET_MONITOR_MAINNET_RPC_DIAL_TIMEOUT=5s

// Prometheus metric endpoint.
NEOFS_NET_MONITOR_METRICS_ENDPOINT=:16512

// Interval between NeoFS metric scrapping.
NEOFS_NET_MONITOR_METRICS_INTERVAL=15m
```

To download actual LOCODE database run `$ make locode`.
Visit LOCODE [repository](https://github.com/nspcc-dev/neofs-locode-db) for more information.

```
// Optional path to NeoFS LOCODE database.
NEOFS_NET_MONITOR_LOCODE_DB_PATH=path/to/database
``` 

## Connect to neofs-dev-env

After `Jebudo` release monitor can be attached to 
[neofs-dev-env](https://github.com/nspcc-dev/neofs-dev-env). Go to 
`docker/docker-compose.devenv.yml` file, make sure that contract script
hashes are correct, specify image version for neofs-net-monitor and then
run `make up-devenv` and `make down-devenv` commands.