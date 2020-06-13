# How to build


## Test 
```yaml
make promu
make
git diff --exit-code
rm -v aws_cloudwatch_exporter
```

## Build 
```yaml
make promu
promu crossbuild -v
```

## Publish 
```yaml
make promu
promu crossbuild -v
make docker DOCKER_REPO=docker.io/slashdevops
docker images
# docker login -u $DOCKER_LOGIN -p $DOCKER_PASSWORD hub.docker.com
make docker-publish DOCKER_REPO=docker.io/slashdevops
make docker-manifest DOCKER_REPO=docker.io/slashdevops

```
