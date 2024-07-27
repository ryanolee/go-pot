LOCAL_DEV_IMAGE_NAME = "go-pot-local-dev"
build:
	goreleaser --snapshot --skip-publish --rm-dist

dev:
	docker build -t $(LOCAL_DEV_IMAGE_NAME) --target=dev .
	docker run -it --rm -v $(PWD):/app $(LOCAL_DEV_IMAGE_NAME) 