# NeoFS Network Monitor

Testnet network monitor. Application scrapes data from network map contract in
neo side chain. Then exposes it to prometheus instance. Then grafana visualise 
data in pretty format. Geo data received from geo ip service `ipstack.com`.

For internal usage.

## How to use 

1. Build image of neofs-network-monitor app.

```
$ make build
...
Successfully built 22b63620bc9d
Successfully tagged neofs-net-monitor:0.1.0
```

2. Specify neofs-net-monitor image version and `ipstack.com` access token in 
   `docker/docker-compose.yml`.

```yml
      - NEOFS_NET_MONITOR_GEOIP_ACCESS_KEY=deabeaf1234567890c0ffecafe080894
```

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
// Script hash of netmap contract in NeoFS sidechain.
NEOFS_NET_MONITOR_CONTRACTS_NETMAP=7b383bc5a385859469f366b08b04b4fcd9a41f55 

// WIF for the NEO client. If not set, it is randomly generated at startup.
NEOFS_NET_MONITOR_KEY=KyH4ASQ1tmm7q9eQKiSzCSH6kxNVbUe3B41EeLaJ15UoMwgZw3Zk 

// Sidechain NEO RPC related configuration.
NEOFS_NET_MONITOR_RPC_ENDPOINT=http://rpc1-morph.preview4.nspcc.ru:24333
NEOFS_NET_MONITOR_RPC_DIAL_TIMEOUT=5s

// Prometheus metric endpoint.
NEOFS_NET_MONITOR_METRICS_ENDPOINT=:16512

// Interval between NeoFS metric scrapping.
NEOFS_NET_MONITOR_METRICS_INTERVAL=15m

// GeoIP ipstack.com related configuration values.
NEOFS_NET_MONITOR_GEOIP_ENDPOINT=http://api.ipstack.com
NEOFS_NET_MONITOR_GEOIP_DIAL_TIMEOUT=5s
NEOFS_NET_MONITOR_GEOIP_ACCESS_KEY=deabeaf1234567890c0ffecafe080894
```