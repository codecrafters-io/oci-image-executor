create_test_image_and_artifacts:
	docker build -t test-image -f Dockerfile .
	docker export $(shell docker create test)
