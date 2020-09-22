DIR=deployment/docker
COMPOSE=${DIR}/docker-compose.yaml

DIND_PREFIX ?= $(HOME)

PREFIX=$(shell echo $(PWD) | sed -e s:$(HOME):$(DIND_PREFIX):)

include .env
export $(shell sed 's/=.*//' .env)

UID=$(shell whoami) 

ifeq ($(CACHE_PREFIX),)
	CACHE_PREFIX=/tmp
endif

test: 
	docker run \
		--network api_default \
		-v $(CACHE_PREFIX)/cache/go:/go/pkg/mod \
		-v $(CACHE_PREFIX)/cache/apk:/etc/apk/cache \
		-v $(PREFIX)/deployment/docker/build:/build \
		-v $(PREFIX)/:/src \
		-v $(PREFIX)/migrations:/migrations \
		-v $(PREFIX)/scripts/test.sh:/test.sh \
		-e UID=$(UID) \
		golang:1.13-alpine /test.sh 

network: 
	docker network create -d bridge api_default; /bin/true

migrate: 
	docker run --network api_default -v `pwd`/migrations:/migrations migrate/migrate:v4.10.0 -source file://migrations -database 'postgres://${DB_USER}:${DB_PASS}@database:5432/testdb?sslmode=disable' drop
	docker run --network api_default -v `pwd`/migrations:/migrations migrate/migrate:v4.10.0 -source file://migrations -database 'postgres://${DB_USER}:${DB_PASS}@database:5432/testdb?sslmode=disable' up
prepare: network
	docker-compose -f ${COMPOSE} -p api up -d --force-recreate

build: 
	docker run -v $(CACHE_PREFIX)/cache/go:/go/pkg/mod \
		-v $(CACHE_PREFIX)/cache/apk:/etc/apk/cache \
		-v $(PREFIX)/deployment/docker/build:/build \
		-v $(PREFIX)/scripts/build.sh:/build.sh \
		-v $(PREFIX)/:/src \
		-v $(PREFIX)/cmd:/src/cmd \
		golang:1.13-alpine /build.sh 

run-api: build
	docker run --network api_default -p 8000:8000 --env-file .env -v `pwd`/deployment/docker/build/api:/api alpine /api

build-worker:
	docker run -v $(CACHE_PREFIX)/cache/go:/go/pkg/mod \
		-v $(CACHE_PREFIX)/cache/apk:/etc/apk/cache \
		-v $(PREFIX)/deployment/docker/build:/build \
		-v $(PREFIX)/scripts/build.worker.sh:/build.sh \
		-v $(PREFIX)/:/src \
		-v $(PREFIX)/cmd:/src/cmd \
		golang:1.13-alpine /build.sh  

build-dispatcher:
	docker run -v $(CACHE_PREFIX)/cache/go:/go/pkg/mod \
		-v $(CACHE_PREFIX)/cache/apk:/etc/apk/cache \
		-v $(PREFIX)/deployment/docker/build:/build \
		-v $(PREFIX)/scripts/build.dispatcher.sh:/build.sh \
		-v $(PREFIX)/:/src \
		-v $(PREFIX)/cmd:/src/cmd \
		golang:1.13-alpine /build.sh  

run-worker: build-worker
	docker run -d --network api_default --name=worker --env-file .env -v `pwd`/deployment/docker/build/worker:/worker alpine /worker

run-dispatcher: build-dispatcher
	docker run -d --network api_default --name=dispatcher --env-file .env -v `pwd`/deployment/docker/build/dispatcher:/dispatcher alpine /dispatcher

clean-worker: 
	docker stop worker
	docker rm worker

clean-dispatcher: 
	docker stop dispatcher
	docker rm dispatcher

swagger: gen-only
	docker run -v --rm -it -v $(PREFIX):/work -w /work quay.io/goswagger/swagger generate spec -m -o swagger.json
