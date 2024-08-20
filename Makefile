VERSION_TAG ?= v0.0.4-dev
REGISTRY_ADDR ?= ccr.ccs.tencentyun.com/public-proxy/deploy-controller

build:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o deploy-controller main.go
## you should handler ccr login first
build-image: build
	docker build --platform linux/amd64 -t $(REGISTRY_ADDR):$(VERSION_TAG) .
	docker push $(REGISTRY_ADDR):$(VERSION_TAG)
deploy-k8s: build-image
	kustomize build kustomize/overlays/dev/ | sed -e "s|VERSION_TAG|${VERSION_TAG}|g" -e "s|REGISTRY_ADDR|${REGISTRY_ADDR}|g" | kubectl apply -f -
re-deploy-k8s:
	kustomize build kustomize/overlays/dev/ | sed -e "s|VERSION_TAG|${VERSION_TAG}|g" -e "s|REGISTRY_ADDR|${REGISTRY_ADDR}|g" | kubectl apply -f -
clean:
	rm -f deploy-controller