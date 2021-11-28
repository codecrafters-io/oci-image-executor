# test-image-executor

Takes a filesystem archive from [docker export](https://docs.docker.com/engine/reference/commandline/export/) and 
an [OCI image config](https://github.com/opencontainers/image-spec/blob/main/config.md), and executes the image 
using firecracker.

Stdout logs are captured & exit code is relayed.

# Interface

test-image-executor --image-tar image.tar --image-config image.json --volumes /var/opt/tools/tester

# Developing Locally

Many of the scripts in this repository aren't customized to work on MacOS, so we use Vagrant to test this locally.

1. Create your Vagrant VM:

```shell
vagrant up
```

2. SSH into the VM and run tests using a sample configuration (defined in `make test_local`):

```shell
vagrant ssh
cd /var/opt/test-image-executor
GITHUB_TOKEN="xxx" make test_local # in the vagrant shell
```
