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

## Connect to neofs-dev-env

After `Jebudo` release monitor can be attached to 
[neofs-dev-env](https://github.com/nspcc-dev/neofs-dev-env). Go to 
`docker/docker-compose.devenv.yml` file, make sure that NeoFS contract script
hash is correct, and run `make up-devenv` command.
