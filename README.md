# NeoFS Network Monitor

Testnet network monitor. Application scrapes data from network map contract in
neo side chain. Then exposes it to prometheus instance. Then grafana visualise 
data in pretty format. Geo data received from geo ip service `ipstack.com`.

For internal usage.

## How to use 

1. Build image of neofs-network-monitor app

```
$ make build
...
Successfully built 22b63620bc9d
Successfully tagged neofs-net-monitor:0.1.0
```

2. Specify `ipstack.com` access token in `docker/docker-compose.yml`

```yml
      - NEOFS_NET_MONITOR_GEOIP_ACCESS_KEY=deadeaf1234567890c0ffecafe080894
```

3. Start environment

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