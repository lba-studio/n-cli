# N-CLI - Send notifications to yourself

Why stare at your laptop when you can go make yourself a coffee and have your computer let you know when it's done compiling that monstrous 4GB monorepo?

# Features

- Desktop notification
- Discord notification through webhooks
- [Planned] Mobile app notification through our mobile app
- [Planned] Slack notification

Do open an issue if you're interested in a notification channel being implemented.

# Installation

Through `go install` (easiest if you have Go installed):

```shell
go install github.com/lba-studio/n-cli@latest
```

If you don't have Go installed: download the latest release for your machine here: [Releases](https://github.com/lba-studio/n-cli/releases/)

# Usage

```sh
n-cli --help

# send "My message here" to your configured destinations. if you haven't configured n-cli, we'll setup a config for you
n-cli send My message here
n-cli s My message here

# where is your config?
n-cli where config

# example - make build takes 20 minutes to complete
make build; n-cli s "Build is done, stop making coffee"

# pro tip: you can set an alias to make the whole command shorter
alias n="n-cli s"
make build; n Build is done;
```

# Configuration

Default config is at `~/.n-cli/config.yaml`

```yaml
discord: # if missing, n-cli won't use Discord as a notification channel
  webhookUrl: https://discord.com/api/webhooks/{yourwebhookurlhere}
  messageFormat: "<@1234> {{message}}"
```
