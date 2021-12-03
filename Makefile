create_test_image:
	docker build -t test-image -f Dockerfile .
	docker export -o image.tar $(shell docker create test-image)

test_executor:
	sudo rm -rf /tmp/firecracker.socket && go build -o main *.go && sudo ./main -image-tar=image.tar -image-config=image-config.json
