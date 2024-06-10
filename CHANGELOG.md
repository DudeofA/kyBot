# Changelog

All notable changes will be documented in this file.

## [6.0] - 6/9/24

## Removed

- All things Wordle

## Added

- Ready Check command much like Dota 2

## [5.1.3] - 4/22/22

## Fix

- New day not reseting statusboard

## [5.1.2] - 4/22/22

- Fixed Scoreboard not updating corrected on each submission

## [5.1.1] - 4/20/22

- Fixed User stats not refreshing after each submission

## [5.1.0] - 3/23/22

## Add

- Persistant Wordle Statistics per user
- Convert Leaderboard to Statusboard
- Add Reminders into Statusboard
- Convert all wordlestats to uints

## [5.0.5] - 3/15/22

- Add State management/caching

## [5.0.4] - 3/5/22

- Fixed new users were not appearing on the scoreboard on their 2nd and on submissions

## [5.0.3] - 3/2/22

- Fixed reminders going out when someone already did their Wordle :)
- Fixed Bot not sending out new wordle message at 12am

## [5.0.2] - 3/1/22

- Fix Leaderboard not using correct date to calculate if today's Wordle is complete

## [5.0.1] - 3/1/22

- Better leaderboard with today's completion status
- Update leaderboard on each wordle submission

## [5.0.0] - 2/23/22

### Structure Change

- All files consolidated to one package
- Finished restructure

## [4.2.0] - 2/23/22

### DB

- Restructured how data is organized, removing many to many relationships

## [4.1.8] - 2/13/22

- Add function to import Wordle stats
- More clarity when Adding Wordle stats fails

## [4.1.7] - 2/8/22

- Add Stat for the person who has the worst first word

## [4.1.6] - 2/6/22

- Add basic stats to the Wordle reminder

## [4.1.5] - 2/3/22

- Move users module into status
- Rename status to component

## [4.1.4] - 2/1/22

- Upon start, bot will look for any wordle games it has missed
- Command to force bot to re-search channel for Wordle games

## [4.1.3] - 2/1/22

- Add changelog to final build so it can be read for version number in Docker

- Automatically removes unused commands upon startup
- In debug mode, also automatically global commands

## [4.1.2] - 2/1/22

- Status that includes version from changelog

## [4.1.1] - 1/31/22

- Gave startup debug messages clarity on registering commands with Discord

## [4.1.0] - 1/30/22

- Wordle Stat Recording

## [4.0.1] - 1/19/22

- Wordle reminders and opt in

## [4.0.0] - 7/21/21

### Version 4 - The 4th rework

- One can dream it will be the final iteration

## [3.0.12] - 3/13/21

- Added meme command using imgflip api

- Removed unnecessary break statements throughout

## [3.0.11] - 2/14/21

- Member command to add users to config defined role

- Fixed new databases being created using old parameters

## [3.0.10] - 6/29/20

- Minecraft server polling using api.kylixor.com/mc
- Update quotelist command to have 2 columns for better readability

## [3.0.9] - 6/9/20

- Vote ending if submitter votes at all (not just downvote)

## [3.0.8] - 5/20/20

- Vote kicking not reporting errors
- Vote kicking not sending userID correctly

## [3.0.7] - 5/12/20

- Changed watch table to accommodate users
- Users can submit their own quote identifiers

- Fixed dependencies in Makefile

## [3.0.6] - 5/11/20

- kick command to votekick users (moves them to the AFK channel)
- Logging submitter of votes, they can now cancel their own votes easily
- Word of the day (maybe it will get real words someday)

- Fixed when sending just the prefix made the bot reply
- KDB version control

## [3.0.5] - 4/11/20

- ~~Minecraft command to poll configured MC servers~~ On hold, moved to website API (which doesn't exist yet)
- Version is pulled from changelog, no need to update readme for version
- Moved dailies to their own function in users.go

- Help not reading README correctly
- Quotes not completing correctly
- Votes not looking as they are supposed to

## [3.0.4] - 4/11/20

- New watch table for monitoring messages (i.e. hangman, votes)
- Added vote command
- Made quote command use vote command

- Improved vote functionality to use watch table

## [3.0.3] - 2/11/20

- Minimum votes needed changed to 3

- Downvoting own vote end the vote

## [3.0.2] - 2/8/20

- Countdown until dailies are up

- Compensation admin function

## [3.0.1] - 2/8/20

- Fixed game starting when changing hangman channel
- Fixed quotelist working as intended
- Changed scheduler to hopefully fix dailies

## [3.0.0] - 11/20/19

### Major Change

- Dropped MongoDB because it was confusing, moving to MySQL

## [2.0.0] - 8/30/19

- MongoDB as backend database instead of JSON file

## [1.2.8] - 8/18/19

- Balanced hangman
- Fix slots displaying letters instead of emojis

## [1.2.1] - 7/16/19

- Hangman!

## [1.1.7] - 7/6/19

- Slots/gambing

## [1.1.2] - 7/1/19

- New 'database' access methods
- Pass objects in 'database' by address instead of writing back changes
- Restructured database (Users are global, servers are individual)

## [1.0.4] - 6/23/19

- Querying age of users/channels/guilds

## [1.0.0] - 5/28/19

- Rewrite bot completely to make future development better and improve programming skills
- Daily coins collection and admin tools
- Allow users to change Discord voice servers
- Reworked quote functions
- Logging
- Account querying

## [0.0.1] - 4/28/2018

- Original code before versioning
