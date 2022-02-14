**Kylixor Multi-Purpose Discord Bot**
=====================================

[![CodeFactor](https://www.codefactor.io/repository/github/dudeofa/kyBot/badge)](https://www.codefactor.io/repository/github/dudeofa/kyBot)  [![Go report](http://goreportcard.com/badge/dudeofa/kyBot)](http://goreportcard.com/report/dudeofa/kyBot)

Dependencies:
~~~
make
go
~~~

Installation steps:
~~~
make
./bin/kylixor
Fill out conf.json file with necessary information
~~~

Commands
--------

Begin help command here:

| Command      | Alias | Description
| -----------  | ----- | -----------
| account      | acc   | Displays how many coins the user has
| age          | age   | Display the create date of the user or Snowflake ID
| dailies      | day   | Get daily coins/credits
| darling      | 02    | Displays a gif of pure happiness
| hangman      | hm    | Interacts with the current hangman game
| help         | h     | Print out this file
| gamble       | slots | Gambles the given credits
| karma        | karma | Displays the current level of karma for the bot
| kick         | k     | Starts a vote to kick the mentioned user out of the voice channel
| minecraft    | mc    | Displays status of currently configured Minecraft servers or given hostname
| ping         | ping  | Send a ping to the bot to measure latency
| quote        | q     | Starts a vote to save a memorable quote
| quotelist    | ql    | Lists the specified quote (i.e. quotelist <identifier>)
| quoterandom  | qr    | Print a random quote from the database
| version      | v     | Gets bot's current version
| voiceserver  | vs    | Change the voice server region
| vote         | poll  | Start a vote/poll using "|" as a separator, i.e. !vote red | yellow)   
| wotd         | word  | Displays the word of the day (CANOODLE)  
 
| Admin CMD    | Alias | Description
| -----------  | ----- | -----------
| compensation | comp  | Gives compensation coins for dailies missed and downtime
| config       | c     | Manipulate the config file
| ip           | ip    | Display external IP of bot
| test         | test  | test
