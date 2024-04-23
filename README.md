# Tony - Discord Bot

>  17th April 2024

The Aussie BroadWAN has its own Discord bot for it's server. This is open for 
development by members of the small community. This is written in [Go] for no
particular reason than to just improve skills in the language. The bot supports
[App Commands] and channel message "moderation". The Tony framework can be 
extended upon if needed for other kind of bot functionalities.


## How to Run

To deploy your own instance of Tony, start by creating a Discord bot 
application. Detailed instructions are available in the [Discord Dev Doc]. 
After setting up your bot, generate a Bot Token and save it in a `.env` file.

> **Note:** Use the provided .env.example as a template. Simply rename it to 
>           `.env` and update its contents accordingly.

Once your bot is configured and added to a server, you're ready to compile and 
run the code. Although future releases might include precompiled binaries, you 
currently need to compile the bot manually. Ensure you have `go` installed by 
consulting the [Go Install] documentation. Then, execute the following commands 
within the project's root directory:

```bash
# Build and Compile the program
go build .

# Run the Program
./tony
```

> **Note:** Remember to load the `.env` file into your environment variables 
>           using a command like `export $(cat .env)`.

### Running Locally with Docker

The instructions below outline how to set up a local environment resembling the 
production setup:

```bash
# Set up a Local Postgres Database
docker network create tony-network
docker pull postgres:latest
docker volume create pgdata
docker run --name postgres                                                     \
    -e POSTGRES_DB=tony                                                        \
    -e POSTGRES_USER=user                                                      \
    -e POSTGRES_PASSWORD=password                                              \
    --network tony-network                                                     \
    -v pgdata:/var/lib/postgresql/data                                         \
    -d --restart unless-stopped postgres:latest

# Build the Project (required after every update)
docker build -t tony .
sudo docker run                                                                \
    --env-file .env                                                            \
    --network tony-network                                                     \
    tony                                                                       
```

Ensure your `.env` file includes the necessary credentials:

```bash
DISCORD_TOKEN=
DISCORD_SERVER_ID=
DISCORD_STARTUP_CHANNEL=tony-dev

DB_HOST=postgres
DB_NAME=tony
DB_USER=user
DB_PASSWORD=password
```

> **Note:** Make sure to populate the `DISCORD_TOKEN` and `DISCORD_SERVER_ID` 
>           fields with your specific bot details.


[Go]: https://go.dev/
[App Commands]: https://discord.com/developers/docs/interactions/application-commands
[Discord Dev Doc]: https://discord.com/developers/docs/getting-started
[Go Install]: https://go.dev/doc/install
