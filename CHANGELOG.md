# Changelog

All notable changes will be documented in this file.

## [5.1.0] - 3/23/22

## Add

-   Persistant Wordle Statistics per user
-   Convert Leaderboard to Statusboard
-   Add Reminders into Statusboard
-   Convert all wordlestats to uints

## [5.0.5] - 3/15/22

## Add

-   Add State management/caching

## [5.0.4] - 3/5/22

## Fix

-   New users were not appearing on the scoreboard on their 2nd and on submissions

## [5.0.3] - 3/2/22

## Fix

-   Reminders going out when someone already did their Wordle :)
-   Bot not sending out new wordle message at 12am

## [5.0.2] - 3/1/22

## Fix

-   Fix Leaderboard not using correct date to calculate if today's Wordle is complete

## [5.0.1] - 3/1/22

## Add

-   Better leaderboard with today's completion status
-   Update leaderboard on each wordle submission

## [5.0.0] - 2/23/22

#### Structure Change

-   All files consolidated to one package
-   Finished restructure

## [4.2.0] - 2/23/22

#### DB

-   Restructured how data is organized, removing many to many relationships

## [4.1.8] - 2/13/22

#### Add

-   Function to import Wordle stats
-   More clarity when Adding Wordle stats fails

## [4.1.7] - 2/8/22

#### Add

-   Add Stat for the person who has the worst first word

## [4.1.6] - 2/6/22

#### Add

-   Add basic stats to the Wordle reminder

## [4.1.5] - 2/3/22

#### Change

-   Move users module into status
-   Rename status to component

## [4.1.4] - 2/1/22

#### Add

-   Upon start, bot will look for any wordle games it has missed
-   Command to force bot to re-search channel for Wordle games

## [4.1.3] - 2/1/22

#### Fix

-   Add changelog to final build so it can be read for version number in Docker

#### Add

-   Automatically removes unused commands upon startup
-   In debug mode, also automatically global commands

## [4.1.2] - 2/1/22

#### Add

-   Status that includes version from changelog

## [4.1.1] - 1/31/22

#### Fix

-   Gave startup debug messages clarity on registering commands with Discord

## [4.1.0] - 1/30/22

#### Add

-   Wordle Stat Recording

## [4.0.1] - 1/19/22

#### Add

-   Wordle reminders and opt in

## [4.0.0] - 7/21/21

#### Version 4 - The 4th rework

-   One can dream it will be the final iteration

## [3.0.12] - 3/13/21

#### Add

-   meme command using imgflip api

#### Fixed

-   Unnecessary break statements throughout

## [3.0.11] - 2/14/21

#### Add

-   Member command to add users to config defined role

#### Fixed

-   Fixed new databases being created using old parameters

## [3.0.10] - 6/29/20

#### Add

-   Minecraft server polling using api.kylixor.com/mc
-   Update quotelist command to have 2 columns for better readability

## [3.0.9] - 6/9/20

#### Fixed

-   Vote ending if submitter votes at all (not just downvote)

## [3.0.8] - 5/20/20

#### Fixed

-   Vote kicking not reporting errors
-   Vote kicking not sending userID correctly

## [3.0.7] - 5/12/20

#### Add

-   Changed watch table to accommodate users
-   Users can submit their own quote identifiers

#### Fixed

-   Dependencies in Makefile

## [3.0.6] - 5/11/20

#### Add

-   kick command to votekick users (moves them to the AFK channel)
-   Logging submitter of votes, they can now cancel their own votes easily
-   Word of the day (maybe it will get real words someday)

#### Fixed

-   Sending just the prefix made the bot reply
-   KDB version control

## [3.0.5] - 4/11/20

#### Add

-   ~~Minecraft command to poll configured MC servers~~ On hold, moved to website API (which doesn't exist yet)
-   Version is pulled from changelog, no need to update readme for version
-   Moved dailies to their own function in users.go

#### Fixed

-   Help not reading README correctly
-   Quotes not completing correctly
-   Votes not looking as they are supposed to

## [3.0.4] - 4/11/20

#### Add

-   New watch table for monitoring messages (i.e. hangman, votes)
-   Added vote command
-   Made quote command use vote command

#### Fix

-   Improved vote functionality to use watch table

## [3.0.3] - 2/11/20

#### Fix

-   Minimum votes needed changed to 3

#### Add

-   Downvoting own vote end the vote

## [3.0.2] - 2/8/20

#### Fix

-   Countdown until dailies are up

#### Add

-   Compensation admin function

## [3.0.1] - 2/8/20

#### Fix

-   Fixed game starting when changing hangman channel
-   Fixed quotelist working as intended
-   Changed scheduler to hopefully fix dailies

## [3.0.0] - 11/20/19

#### Major Change

-   Dropped MongoDB because it was confusing, moving to MySQL

## [2.0.0] - 8/30/19

#### Added

-   MongoDB as backend database instead of JSON file

## [1.2.8] - 8/18/19

#### Fixed

-   Balanced hangman
-   Fix slots displaying letters instead of emojis

## [1.2.1] - 7/16/19

#### Added

-   Hangman!

## [1.1.7] - 7/6/19

#### Added

-   Slots/gambing

## [1.1.2] - 7/1/19

#### Added

-   New 'database' access methods
-   Pass objects in 'database' by address instead of writing back changes
-   Restructured database (Users are global, servers are individual)

## [1.0.4] - 6/23/19

#### Added

-   Querying age of users/channels/guilds

## [1.0.0] - 5/28/19

#### Added

-   Rewrite bot completely to make future development better and improve programming skills
-   Daily coins collection and admin tools
-   Allow users to change Discord voice servers
-   Reworked quote functions
-   Logging
-   Account querying

## [0.0.1] - 4/28/2018

#### Added

-   Original code before versioning
