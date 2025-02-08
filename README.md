# Pal

`pal` is an AI assistant for terminals. The philosophy here is to feel like a classic shell utility with a bit of magic.

![demo3](https://github.com/user-attachments/assets/c8a9356d-0bb9-43f3-9003-de8b71329674)

For now, `fish` and `zsh` are fully supported and testing is done on Linux only. MacOS might just work too.

Perhaps unsurprisingly, xkcd [has](https://xkcd.com/1168/) elucidated the core situation inspiring this software:

<br><img width="713" alt="image" src="https://github.com/user-attachments/assets/93f58393-2de7-466e-ba30-9fa2e32635af" />

<br>`pal` helps diffuse the bomb that nukes our focus when leaving the shell to surf for answers on the web.

## Quickstart

It's a single binary that you can just download into any location on your `$PATH`:

```sh
wget https://github.com/scottyeager/Pal/releases/latest/download/pal-linux-amd64 -O /usr/local/bin/pal
chmod +x /usr/local/bin/pal
```

If the command name `pal` is already taken on your system, feel free to give it a name of your choice.

### Config

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
# Config saved successfully at ~/.config/pal_helper/config.yaml
```

### Abbreviations

Abbreviations are an optional feature of `pal` that are highly recommended. When they are enabled, you can autofill the contents of the suggestions from the last `pal` invocation like this:

```sh
pal1 # Hit space and first suggestion will be filled
pal2 # Etc
```

Both `fish` and `zsh` are supported for abbreviations.

#### fish

To activate abbreviations for `fish` add the following to your `~/.config/fish/config.fish`:

```sh
pal --fish-abbr | source
```

#### zsh

On `zsh` you will need the [zsh-abbr](https://github.com/olets/zsh-abbr) plugin. Broadly, there are two ways to install it:

1. Pal contains a copy of zsh-abbr and you can install it by adding this to `~/.zshrc`:

```sh
source $(pal --zsh-abbr)
```

2. If you prefer, [install `zsh-abbr`](https://zsh-abbr.olets.dev/installation.html) in one of the usual ways, such as with a `zsh` plugin manager or a system package

Abbreviations for `zsh` must also be enabled in the `pal` config. You can run `pal /config` again or edit `~/.config/pal_helper/config.yaml` if you didn't enable them initially.

> Note: I'm a `fish` user, and I added `zsh` support after a bit of research into how to provide a similar experience. If you have ideas for how to make the `zsh` integration better or how to add support for your favorite shell, please open an issue on this repo and let me know. Thanks!

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

The `/models` command is used to select models and also check the currently selected model:

```
pal /models
```

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
