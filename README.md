# ✉️ Notification CLI (n-cli) - Send notifications to yourself through the command line

Why stare at your laptop when you can go make yourself a coffee and have your computer let you know when it's done compiling that monstrous 4GB monorepo?

# 🚀 Features

- Works out-of-the-box - no need to go through any external service (other than your chat apps, of course)
- Desktop notification
- Discord notification through [Discord webhooks](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks)
- Slack notification through [Slack workflow webhooks](https://slack.com/intl/en-gb/help/articles/360041352714-Create-workflows-that-start-with-a-webhook)
- [Planned] Mobile app notification through our mobile app

Do open an issue if you're interested in a notification channel being implemented.

# 👨🏻‍💻 Installation

Through `go install` (easiest if you have Go 1.21 installed):

```shell
go install github.com/lba-studio/n-cli@latest
```

If you don't have Go installed: download the latest release for your machine here: [Releases](https://github.com/lba-studio/n-cli/releases/)

## Which one should I download?

{version} refers to the current version of n-cli.

| MacOS Apple Silicon                 | MacOS Intel                         | Windows Intel 64-bit                 |
| ----------------------------------- | ----------------------------------- | ------------------------------------ |
| n-cli-{version}-darwin-arm64.tar.gz | n-cli-{version}-darwin-amd64.tar.gz | n-cli-{version}-windows-amd64.tar.gz |

_\*Open an issue / PR if this table is wrong, thank you!_

# 🐈 Usage

```sh
n-cli --help

# send "My message here" to your configured destinations. if you haven't configured n-cli, we'll setup a config for you
n-cli send My message here
n-cli s My message here

# example - make build takes 20 minutes to complete
make build; n-cli s "Build is done, stop making coffee"

# alternatively, run your shell command through n-cli
n-cli run make build

# make sure to use `--` if you'd like to pass in flags
n-cli r -- make build --whatever-args-i-have-here

# pro tip: you can set an alias to make the whole command shorter
alias n="n-cli s"
make build; n Build is done;

# useful commands
n-cli init # optional: initializes & configures n-cli without running anything
n-cli where config # where is your config?
n-cli version # get version
```

# 📝 Configuration

Default config is at `~/.n-cli/config.yaml`

```yaml
discord: # if missing, n-cli won't use Discord as a notification channel
  # https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks
  webhookUrl: https://discord.com/api/webhooks/{yourwebhookurlhere} # required
  messageFormat: "<@1234> {{message}}" # optional

slack: # if missing, n-cli won't use Slack as a notification channel
  # you can create one by following the steps here https://slack.com/intl/en-gb/help/articles/360041352714-Create-workflows-that-start-with-a-webhook
  # must have "message" as a variable. sample payload to the webhook: { "message": "{{message}}" }
  webhookUrl: https://hooks.slack.com/triggers/ABCDEFG123/123456789/whateverstringishere # required
  messageFormat: "{{message}}" # optional
```
