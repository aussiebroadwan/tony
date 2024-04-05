# Tony - Discord Bot

>  2nd April 2024

The Aussie BroadWAN has its own Discord bot for it's server. This is open for 
development by members of the small community. This is written in [Go] for no
particular reason than to just improve skills in the language. The bot supports
[App Commands] and channel message "moderation". The Tony framework can be 
extended upon if needed for other kind of bot functionalities.


## How to Run

The first thing you will need to do is create a bot application on discord. This
can be done by following the [Discord Dev Doc]. Once you have made a bot, create
or fetch you Bot Token and put it in a `.env` file.

> **Note:** There is an example `.env` file called `.env.example`, you can 
>           rename this in your directory to `.env` and use that.

Once you have setup your app, bot and added it into a discord server, then you 
are ready to run the program. Eventually there will be binaries avaliable but
for now you can just compile it yourself. You will need `go` installed, if you
don't then go to the [Go Install] docs. Once you have the `go` then you can run
the following in the project directory:

```bash
# Build and Compile the program
go build .

# Run the Program
./tony
```

> **Note:** Future development will move the program into a Docker container for
>           easier deploying, but this is just a rush job to get a basic bot up 
>           and running.

## Current Bot Features

- `ping`: 
    Sends the user a `Pong @<user>!` message. This is only for testing.

    - `ping button`: Responses with a button which the user can press for the
        same functionality as the `ping` command

- `remind`:
    A system to add deplayed message or reminders for users.

    - `add <time> <message>`: The message to add and when to remind the user
    - `del <id>`: Deletes a message, assuming the user owns the message
    - `status <id>`: Get how much time is left on a reminder
    - `list`: List the ID and times of all the user's reminders
  
- `tech-news` [**MODERATION**]:
    A system to ensure posts being made in the `#tech-news` channel is in a 
    specific format.

- `rss` [**MODERATION**]:
    A system to ensure posts being made in the `#rss` channel is in a specifc 
    format.

- `autopin` [**REACTION**]:
    A system which automatically pins messages if there are 5 or more 📌 reacts.

[Go]: https://go.dev/
[App Commands]: https://discord.com/developers/docs/interactions/application-commands
[Discord Dev Doc]: https://discord.com/developers/docs/getting-started
[Go Install]: https://go.dev/doc/install