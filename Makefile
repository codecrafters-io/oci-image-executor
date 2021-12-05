create_test_image:
	docker build -t test-image -f Dockerfile .
	docker export -o image.tar $(shell docker create test-image)

create_redis_image:
	docker export -o image.tar $(shell docker create redis:latest)

test_executor:
	sudo rm -rf /tmp/firecracker.socket && go build -o main *.go && sudo ./main -image-tar=image.tar -image-config=image-config.json -volumes /tmp:/var/opt/mounted-dir
