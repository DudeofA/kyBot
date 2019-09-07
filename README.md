**Kylixor Multi-Purpose Discord Bot**
v2.0.0
=====================================

Dependencies:
~~~
make
go
github.com/bwmarrin/discordgo
github.com/bwmarrin/dgvoice
github.com/jasonlvhit/gocron
go.mongodb.org/mongo-driver/mongo
~~~

Installation steps:
~~~
make
./bin/kylixor
~~~
Fill out conf.json file with necessary information

Commands
--------

| Command     | Alias | Description
| ----------- | ----- | -----------
| account     | acc   | Displays how many coins the user has
| age         | age   | Display the create date of the user or Snowflake ID
| dailies     | day   | Get daily coins/credits
| darling     | 02    | Displays a gif of pure happiness
| hangman     | hm    | Interacts with the current hangman game
| help        | h     | Print out this file
| gamble      | slots | Gambles the given credits
| karma       | karma | Displays the current level of karma for the bot
| ping        | ping  | Send a ping to the bot to measure latency
| quote       | q     | Starts a vote to save a memorable quote
| quotelist   | ql    | Lists the specified quote (i.e. quotelist 4)
| quoterandom | qr    | Print a random quote from the database
| version     | v     | Gets bot's current version
| voiceserver | vs    | Change the voice server region

| Admin CMD   | Alias | Description
| ----------- | ----- | -----------
| config      | c     | Manipulate the config file
| ip          | ip    | Display external IP of bot
| test        | test  | test
