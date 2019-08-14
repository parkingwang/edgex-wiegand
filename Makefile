BINARY?=$(shell cat name.txt)

# Build opts
BUILD_ENV?=CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH}

IMAGE_VER?=0.5.2
IMAGE_ORG=registry.cn-shenzhen.aliyuncs.com/edge-x

IMAGE_NAME_PLATFORM=${IMAGE_ORG}/${BINARY}:${IMAGE_VER}-${GOOS}_${GOARCH}
IMAGE_NAME_VERSION=${IMAGE_ORG}/${BINARY}:${IMAGE_VER}

IMAGE_NAME_ARM=${IMAGE_ORG}/${BINARY}:${IMAGE_VER}-${GOOS}_arm
IMAGE_NAME_ARM64=${IMAGE_ORG}/${BINARY}:${IMAGE_VER}-${GOOS}_arm64
IMAGE_NAME_AMD64=${IMAGE_ORG}/${BINARY}:${IMAGE_VER}-${GOOS}_amd64


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
	sudo docker build -t $(IMAGE_NAME_PLATFORM) .

# 推送Image到Registry
push:
	@echo ">>> Docker PUSH IMAGE: $<"
	sudo docker push $(IMAGE_NAME_PLATFORM)

# 创建Minifest
manifest:
	@echo ">>> Docker PUSH MANIFEST: $<"
	sudo docker manifest create --amend $(IMAGE_NAME_VERSION) $(IMAGE_NAME_ARM) $(IMAGE_NAME_ARM64) $(IMAGE_NAME_AMD64)
	sudo docker manifest annotate $(IMAGE_NAME_VERSION) $(IMAGE_NAME_ARM) --os linux --arch arm
	sudo docker manifest annotate $(IMAGE_NAME_VERSION) $(IMAGE_NAME_ARM64) --os linux --arch arm64  --variant v8
	sudo docker manifest annotate $(IMAGE_NAME_VERSION) $(IMAGE_NAME_AMD64) --os linux --arch amd64
	sudo docker manifest push $(IMAGE_NAME_VERSION)

clean:
	rm -f $(BINARY)

.PHONY: clean build
