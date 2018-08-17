.PHONY: all
PACKAGE  = k8s-dns-updater
M = $(shell printf "\033[34;1m▶\033[0m")
SRCS = $(shell git ls-files '*.go' '*/*.go' | grep -v '^vendor/')

print-%: ; @echo $*=$($*)

default: build

all: build ; $(info $(M) running all builds...)

lint: ; $(info $(M) running golint...)
	scripts/validate-golint.sh $(SRCS)

prepare: ; $(info $(M) preparing for build...) @ ## Update all versions
	scripts/update_version.sh

build: prepare lint ; $(info $(M) building binary...)
	go build

image: ; $(info $(M) building container image...)
	docker build --no-cache -t $(PACKAGE):$(shell ./k8s-dns-updater version) .

clean: ; $(info $(M) cleaning binary...)
	rm -f $(PACKAGE)