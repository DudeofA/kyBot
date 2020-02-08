all: clean kylixor

compile: ## Just builds
	go build -o bin/kylixor *.go

kylixor: ## Default action. Builds Kylixor.
	go get -u -v "github.com/bwmarrin/discordgo"
	go get -u -v "github.com/bwmarrin/dgvoice"
	go get -u -v "github.com/robfig/cron"
	go get -u -v "github.com/go-sql-driver/mysql"
	
	go build -o bin/kylixor *.go

clean: ## Removes compiled Kylixor binaries.
	rm -f bin/kylixor

test: compile ## Test run the bot
	./bin/kylixor

windows:
	env GOOS="windows" GOARCH="amd64" go build -o bin/kylixor.exe

help: ## Shows this helptext.
	grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
