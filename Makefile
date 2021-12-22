create_test_image:
	docker build -t test-image -f Dockerfile .
	docker export -o image.tar $(shell docker create test-image)

create_redis_image:
	docker export -o image.tar $(shell docker create redis:latest)

test_executor:
	sudo rm -rf /tmp/firecracker.socket && go build -o main ./cmd/ && sudo ./main --image-tar=image.tar --image-config=image-config.json --volume /root/oci-image-executor:/var/opt/mounted-dir

kill_executor:
	kill $$(ps aux | grep firecracker | head -n 2 | tail -n 1 | awk '{print $$2}')

download_kernel:
	mkdir -p /root/firecrafter-resources
	wget https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin -P /root/firecracker-resources/
