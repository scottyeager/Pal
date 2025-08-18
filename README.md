# Pal

`pal` is an AI assistant for terminals. The philosophy here is to feel like a classic shell utility with a bit of magic.

![demo3 1](https://github.com/user-attachments/assets/e6f4be6e-788e-453e-9f27-2b61b76755aa)

`fish` and `zsh` are fully supported on Linux and macOS.

Perhaps unsurprisingly, xkcd [has](https://xkcd.com/1168/) elucidated the core situation inspiring this software:

<br><img width="713" alt="image" src="https://github.com/user-attachments/assets/93f58393-2de7-466e-ba30-9fa2e32635af" />

<br>`pal` helps diffuse the bomb that nukes our focus when leaving the shell to surf for answers on the web.

## Quickstart

It's a single binary that you can download into any location on your `$PATH`:

macOS (Apple Silicon):

```sh
curl -L https://github.com/scottyeager/Pal/releases/latest/download/pal-darwin-arm64 -o /usr/local/bin/pal
chmod +x /usr/local/bin/pal
```

macOS (Intel):

```sh
curl -L https://github.com/scottyeager/Pal/releases/latest/download/pal-darwin-amd64 -o /usr/local/bin/pal
chmod +x /usr/local/bin/pal
```

Linux (x86_64):

```sh
wget https://github.com/scottyeager/Pal/releases/latest/download/pal-linux-amd64 -O /usr/local/bin/pal
chmod +x /usr/local/bin/pal
```

If the command name `pal` is already taken on your system, feel free to give it a name of your choice.

To conveniently install autocompletions and the abbreviation feature:

```sh
# fish
pal --fish-config >> ~/.config/fish/config.fish

# zsh (autocomplete is an optional feature in zsh--see details below)
pal --zsh-config >> ~/.zshrc

# Start a new shell or source your config file from an existing shell
```

More information about these features and how to install them individually can be found at the relevent docs pages: autocompletion and [abbreviations](https://github.com/scottyeager/Pal/blob/main/docs/abbreviations.md).

## Config

You will need to provide an API key for an LLM provider. Several providers listed below have a free tier.

The free tiers sometimes require that you agree to collection and use of the data you submit. Those providers also have paid plans without the data collection requirement. See the links for details.

Supported providers:

* [DeepSeek](https://platform.deepseek.com/)
* [Anthropic](https://console.anthropic.com/)
* [OpenAI](https://platform.openai.com/)
* [Hugging Face Inference API](https://huggingface.co/docs/api-inference/getting-started) (free with [no data collection](https://huggingface.co/docs/api-inference/security), but slow)
* [Mistral](https://console.mistral.ai/) (free with [data collection](https://mistral.ai/terms/#our-free-services))
* [Google](https://ai.google.dev/) (free with [data collection](https://ai.google.dev/gemini-api/terms#unpaid-services))
* [OpenWebUI](openwebui.com) (self hosted models via Ollama, see [guide](https://github.com/scottyeager/Pal/blob/main/docs/openwebui.md))
* Any OpenAI API compatible provider (via manual config)

For interactive configuration, run:

```sh
pal /config
# The command will print where your config was saved (varies by OS)
```

## Abbreviations

Abbreviations are an optional feature of `pal` that are highly recommended. When they are enabled, you can autofill the contents of the suggestions from the last `pal` invocation like this:

```sh
pal1 # Hit space and first suggestion will be filled
pal2 # Etc
```

Both `fish` and `zsh` are supported for abbreviations. If you followed the quickstart, abbreviations will be available in every new shell or after sourcing the shell config file. For more info, see [abbreviations](https://github.com/scottyeager/Pal/blob/main/docs/abbreviations.md).

### Paths

`pal` uses standard OS-specific paths:

- Config file:
  - macOS: `~/Library/Application Support/pal_helper/config.yaml`
  - Linux: `~/.config/pal_helper/config.yaml`
- Abbreviations file (stores suggestions for pal1, pal2, ...):
  - macOS: `~/Library/Application Support/pal_helper/expansions.txt`
  - Linux: `~/.local/share/pal_helper/expansions.txt`

## Autocompletion

Since `pal` is built with [Cobra](https://github.com/spf13/cobra), it's able to generate autocompletions for a variety of shells automatically. Currently only the `fish` and `zsh` completions are exposed.


If you followed the quickstart, then you've already installed the autocompletions. These instructions are for installing the autocompletions separately from the abbreviations feature.

### Fish

To activate autocompletions add the following to your `~/.config/fish/config.fish`:

```sh
pal --fish-completion | source
```

### Zsh

In your `~/.zshrc`:

```sh
# Make sure these lines are already present somewhere
autoload -Uz compinit
compinit

# Add this line to load pal's completions
source <(pal --zsh-completion)
```

## Usage

The basic usage pattern is to write `pal` and then describe or ask about the command you need:

```text
pal Set a static IP for eth0
```

`pal` asks the model to provide a short list of possible commands. If it does, they will be shown.

If you have abbreviations enabled, you can expand the suggestions:

```sh
pal1 # Hit space to expand
```

Sometimes a refusal message might be shown if the model can't or won't provide a command suggestion. You can try again or switch to `/ask` mode to get more information.

### Ask mode

`/ask` mode can be used to pass general queries through to the model, without an expectation that it will suggest shell commands in response.

```sh
pal /ask Why is the sky blue
```

### Special characters

In shells like `fish` and `zsh`, the `?` is reserved for globbing, and this will cause problems if you try to use it without quoting or escaping. Thankfully, it doesn't really matter if you just omit the question mark when asking a question to an LLM. Same goes for appostrophes, which are used for quoting by shells too.

```sh
pal /ask whats the reason LLMs dont need punctuation marks to understand me
```

If necessary, you can pass special characters to `pal` by quoting them:

```sh
pal /ask what does this do: 'ls *.log'
```

As of `fish` version 4, the question mark will no longer be used as a glob character by default. You can enable this behavior in earlier versions of `fish` like this:

```fish
set -U fish_features qmark-noglob
# Takes effect in new shells
```

### Stdin

Anything passed to `pal` on `stdin` will be included with the prompt. That means you can do this:

```sh
cat README.md | pal /ask please summarize
```

Of course, that could also be written like this:

```sh
pal /ask please summarize $(cat README.md)
```

By redirecting `stderr`, error messages can be sent to `pal`:

```sh
apt install ping
# There's no package named "ping" so this is an error

# Zsh and Bash
apt install ping |& pal

# Fish
apt install ping &| pal
```

In this case, the error message is enough for the model to suggest the correct install command. You can also provide additional instructions or context as usual:

```sh
docker ps | pal how can I print the first four characters of the container ids only
```

### Model selection

The `/models` command can be used to view and select from configured models:

```
pal /models
```

To select a model by entering its name (with autcompletion), use `/model`:

```
pal /model mistral/codestral-latest
```

With no argument `/model` prints the currently selected model.

For providers added through interactive config, a default set of models will be included. Depending on the provider, additional models may be available that could be added by editing the config file directly. You can also remove models you don't use so they won't show up in model selection list.

### Git commit

The `/commit` command is used to stage changes in Git repos and automatically generate commit messages:

```
pal /commit
```

It works like this:

1. Any modified files are `git add`ed (new files must be added manually)
2. Diffs for the current commit and ten previous commit messages for context are sent to the LLM
3. The suggested commit message is opened for review and editing if needed

It's possible to abort the commit by deleting the message and saving before exiting the editor.


### Temperature

In the context of LLMs, *temperature* refers to the amount of randomness introduced when generating responses. With temperature of 0, responses are deterministic. With temperature of 2, you are working with an artist.

By default, `pal` uses a hopefully sensible hard coded temperature for the task at hand. You can override the temperature for any command that interacts with the AI backend like this:

```
pal -t 2 /ask write a poem

pal --temperature 0 /commit
```

If you want to set the temperature for command suggestions, use the `/cmd` command explicitly:

```
pal -t 2 /cmd show me a crazy command
```

Without a slash command specified, `pal` will ignore the temperature flag and it will get treated as input for the AI.

## Which models to use?

The command completion function of `pal` works well with the latest generation of flagship models:

* `deepseek-chat` (v3)
* `claude-3-5-sonnet-latest`
* `gpt-4o`

It also works fairly well with less expensive "mini" models:

* `claude-3-5-haiku-latest`
* `gpt-4o-mini`

These might have trouble with some more complex commands though, and the token requirements of command suggestion with `pal` are generally so low that the cost savings might be hard to notice.

If you use the Hugging Face API, then the best choice seems to be `deepseek-ai/DeepSeek-R1-Distill-Qwen-32B` but it can be rather slow.

When it comes to `/ask` mode, the experience will be on par with a regular chat bot conversation. So the flagship models are a good choice and the reasoning models (`o1` and `deepseek-reasoner`) can be handy for more difficult queries.

The reasoning models might be a good choice for especially difficult command suggestion requests too, but so far I didn't find many cases to reach for them.

## Why?

There are other terminal based AI projects, so why build another one? The short answer is that none quite provided the experience I wanted for this particular use case.

One category of in shell assistants drop you into a whole new shell enviroment. This is great for some uses cases, such as the excellent [Aider](https://github.com/Aider-AI/aider) project. But I'm still spending a lot of time in my good ol' shell.

Another category leans a bit more toward a TUI, showing a menu of results on the screen to choose from, for example. The experince I desire is closer to a classic shell utility—like an `ls` that lists ideas for the next command to run, instead of listing files.

Finally, who can miss the chance to build something with the incredibly good and cheap DeepSeek API? Times are a changing, so let's have some fun with it :)
