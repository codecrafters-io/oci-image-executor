# oci-image-executor

Executes an [OCI image](https://github.com/opencontainers/image-spec) using [Firecracker](https://github.com/firecracker-microvm/firecracker).

Logs from the executed process (both stdout and stderr) are sent to stdout. Logs from the executor 
itself are sent to stderr. The exit code matches that of the process.

# Interface

```shell
oci-image-executor \
    --image-tar image.tar \
    --image-config image-firecracker-config.json \
    --volumes /var/user-code-submission:/app,/tools/binary:/your-binary
```

- `--image-tar`: Path to the image tar file, created using [docker export](https://docs.docker.com/engine/reference/commandline/export/)
- `--image-config`: Path to an [OCI image config](https://github.com/opencontainers/image-spec/blob/main/config.md) file
- `--volume`: Copy a directory or file from the host into the VM (changes will not be synced back to the host)

# Developing Locally

Many of the scripts in this repository aren't customized to work on macOS, so we use Vagrant to test this locally.

1. Create your Vagrant VM:

```shell
vagrant up
```

2. SSH into the VM and run tests using a sample image:

```shell
vagrant ssh
cd /var/opt/oci-image-executor
make create_test_image # in the vagrant shell
```
