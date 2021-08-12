IMG ?= idoko/fct:0.1
FCT_BIN ?= ./build/fct

build:
	go build -o ${FCT_BIN} ./cmd/flatcartag

clean:
	rm -rf ./build

docker-build:
	docker build -t ${IMG}

docker-push:
	docker push ${IMG}

install:
	kubectl apply -f ./config

uninstall:
	kubectl delete -f ./config || true