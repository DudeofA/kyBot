all: clean kylixor

kylixor: ## Default action. Builds Kylixor.
	go get -u "github.com/bwmarrin/discordgo"
	go get -u "github.com/bwmarrin/dgvoice"
	go get -u "github.com/jasonlvhit/gocron"
	go get -u "go.mongodb.org/mongo-driver/mongo"
	
	go build -o bin/kylixor *.go

clean: ## Removes compiled Kylixor binaries.
	rm -f kylixor

install: ## Copies kylixor binary to /usr/local/bin for easy execution and restarts the service
	systemctl restart kylixor

help: ## Shows this helptext.
	grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
