# Release instructions

## Pre-release checks

These should run successfully:
* `make bin`;
* `make image`;

## Release steps

1. Update `CHANGELOG.md`, `VERSION` and `docker-compose` files and create PR.


2. After successful PR merge:
   * Create and sign tag: `git tag -s v0.7.1`
   * Push it to origin master: `git push origin master v0.7.1`
   

3. Build and push image to a Docker Hub:
```shell
$ make image
$ docker push nspccdev/neo-exporter:0.7.1
```


4. Make a proper GitHub release page
