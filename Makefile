VERSION = $(shell git describe --tags --always --dirty)
# go tool nm ./bot | grep version #to get the path of the version variable
# https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications
LDFLAGS=-ldflags "-X 'github.com/sheshenia/simplebot/pkg/bot.version=$(VERSION)'"
OSARCH=$(shell go env GOHOSTOS)-$(shell go env GOHOSTARCH)

# prefer to specify token as make env, using with: make run TOKEN=your_bot_token
TOKEN?=your_telegram_bot_token
DEBUG?=true

run:
	go run ./cmd/bot --token=$(TOKEN) --debug=$(DEBUG)

build:
	go build ./cmd/bot

SIMPLEBOT=\
	simplebot-linux-amd64 \
	simplebot-darwin-amd64 \
	simplebot-darwin-arm64 \
	simplebot-windows-amd64

my: simplebot-$(OSARCH)

$(SIMPLEBOT): cmd/bot
	GOOS=$(word 2,$(subst -, ,$@)) GOARCH=$(word 3,$(subst -, ,$(subst .exe,,$@))) go build $(LDFLAGS) -o $@ ./$<

clean:
	rm -rf simplebot-*

simplebot-%-$(VERSION).zip: simplebot-%
	# creating the release folder with GOOS-VERSION
	mkdir simplebot-$(*)-$(VERSION)
	# copy assets to release folder $(*) = % in command
	cp -r ./assets ./simplebot-$(*)-$(VERSION)
	# move binary to release folder
	mv ./simplebot-$(*) ./simplebot-$(*)-$(VERSION)
	# move all release folder content to archive
	cd ./simplebot-$(*)-$(VERSION) && zip -mr ../$@ *
	# delete release folder
	rm -r ./simplebot-$(*)-$(VERSION)

release: \
	simplebot-linux-amd64-$(VERSION).zip \
	simplebot-darwin-amd64-$(VERSION).zip \
	simplebot-darwin-arm64-$(VERSION).zip \
	simplebot-windows-amd64-$(VERSION).zip

# generates json files from media folders in assets
mediajson:
	cd ./assets && ./mediajson.sh


# deploy on Linux CentOS server. All this vars are replaced with arguments
# make deploy IP=some.ip USER=some_user
IP=127.0.0.1
USER=root
ADDR=$(USER)@$(IP)
NAME=simplebot-linux-amd64
PATHNAME=/opt/app/simplebot/$(NAME)

deploy: $(NAME)
	ssh $(ADDR) rm $(PATHNAME) #removes file but doesn't stop the systemctl process https://www.linuxquestions.org/questions/linux-general-1/scp-text-file-busy-error-365198/
	scp $(NAME) $(ADDR):$(PATHNAME) #copy our build to remote machine,without previous step "Text file busy" error
	ssh $(ADDR) systemctl restart $(NAME) #restart systemctl process

.PHONY: my $(SIMPLEBOT) clean release test build run mediajson deploy