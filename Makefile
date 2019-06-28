BINARY?=$(shell cat name.txt)

# Build opts
BUILD_ARCH=$(shell echo ${OSARCH})
BUILD_ENV?=CGO_ENABLED=0 GOOS=linux GOARCH=${BUILD_ARCH}

IMAGE_TAG?=0.0.2-${BUILD_ARCH}
IMAGE_ORG?=registry.cn-shenzhen.aliyuncs.com/edge-x
IMAGE_NAME=${IMAGE_ORG}/${BINARY}:${IMAGE_TAG}


build: clean
	@echo ">>> Go BUILD $(BINARY)"
	@$(BUILD_ENV) go build $(LD_FLAGS) -o $(BINARY).raw .
	rm -f $(BINARY)
	upx -o $(BINARY) $(BINARY).raw
	rm -f $(BINARY).raw


image: _build_image
	@echo ">>> Docker IMAGE: $<"


# 构建Image
_build_image: build
	sudo docker build --build-arg IMAGE=scratch -t $(IMAGE_NAME) .

# 推送Image到Registry
push:
	@echo ">>> Docker PUSH IMAGE: $<"
	sudo docker push $(IMAGE_NAME)

.PHONY: clean build
clean:
	rm -f $(BINARY)