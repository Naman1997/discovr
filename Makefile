include .env

IMAGE := fedora_nmap
CONTAINER_NAME := discovr_nmap
NMAP_WIN_ZIP := nmap-$(NMAP_VERSION)-win32.zip

all: build

build: get_nmap_binary get_nmap_win_zip
	@rm -f discovr
	@go build -ldflags="-X 'github.com/Naman1997/discovr/internal.NmapVersion=$(NMAP_VERSION)'" -v

get_nmap_binary:
ifeq (,$(wildcard assets/nmap))
	@docker buildx build -f nmap.Dockerfile . --tag $(IMAGE) --build-arg FEDORA_PACKAGE=$(FEDORA_PACKAGE)
	@docker create --name $(CONTAINER_NAME) $(IMAGE)
	@docker wait $(CONTAINER_NAME)
	@docker cp $(CONTAINER_NAME):/usr/bin/nmap assets/nmap
	@docker rm -v $(CONTAINER_NAME)
endif
	
get_nmap_win_zip:
ifeq (,$(wildcard assets/$(NMAP_WIN_ZIP)))
	@wget https://nmap.org/dist/$(NMAP_WIN_ZIP) -O assets/$(NMAP_WIN_ZIP)
endif

clean:
	@rm -f discovr assets/$(NMAP_WIN_ZIP) assets/nmap
