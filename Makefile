all: clean kylixor

kylixor: ## Default action. Builds Kylixor.
	@go get "github.com/bwmarrin/discordgo"
	@go get "github.com/bwmarrin/dgvoice"
	@go get "github.com/jasonlvhit/gocron"
	@go build -o bin/kylixor *.go

clean: ## Removes compiled Kylixor binaries.
	@rm -f kylixor

install: ## Copies kylixor binary to /usr/local/bin for easy execution and restarts the service
	@cp -f data/kylixor /usr/local/bin/
	#@systemctl restart kylixor

help: ## Shows this helptext.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
