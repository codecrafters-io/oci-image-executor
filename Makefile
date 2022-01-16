current_version_number := $(shell git tag --list "v*" | sort -V | tail -n 1 | cut -c 2-)
next_version_number := $(shell echo $$(($(current_version_number)+1)))

build:
	go build -o main ./cmd/

release:
	git tag v$(next_version_number)
	git push origin master v$(next_version_number)

create_test_image:
	docker build -t test-image -f Dockerfile .
	docker export -o image.tar $(shell docker create test-image)

create_redis_image:
	docker export -o image.tar $(shell docker create redis:latest)

test_executor: build
	sudo ./main --image-tar=image.tar --image-config=image-config.json --volume $$(pwd):/var/opt/mounted-dir --env TEST=hey --working-dir="/var/opt/mounted-dir" /usr/bin/ls

kill_executor:
	kill $$(ps aux | grep firecracker | head -n 2 | tail -n 1 | awk '{print $$2}')

download_kernel:
	mkdir -p /root/firecrafter-resources
	wget https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin -P /root/firecracker-resources/

test_boot_time: create_test_image
	time make test_executor
