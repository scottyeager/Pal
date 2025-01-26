# Pal

`pal` is an AI assistant for terminals. The philosophy here is to feel like a classic shell utility with a bit of autocomplete magic.

The primary focus will be supporting fish, zsh, and bash on Linux to start. MacOS might just work too.

Perhaps unsurprisingly, xkcd [has](https://xkcd.com/1168/) elucidated the core situation inspiring this software:

<img width="713" alt="image" src="https://github.com/user-attachments/assets/93f58393-2de7-466e-ba30-9fa2e32635af" />

'pal' helps diffuse the bomb that nukes your focus when leaving the shell to surf for answers on the web.

## Quickstart

It's a single binary that you can just download into any location on your `$PATH`:

```
# Show how to wget to /usr/local/bin
```

If the command name `pal` is already taken on your system, feel free to give it a name of your choice.

```
# Show how to wget to /usr/local/bin/hal
```

### Config

You will need to provide an API key for an LLM provider.

Supported providers:

* DeepSeek
* Others coming soon (OpenAI, Anthropic, Ollama, ...?)

For interactive configuration, run:

```
pal /config
# Config file written to ~/.config/pal_helper/config.yaml
```

### Completions

Autocompletion is an optional feature of `pal` that's highly recommended. To install the completions:

```
pal /complete
```

## Usage

The basic usage pattern is to write `pal` and then your quesion:

```
pal How do I set a static IP for eth0?
```

`pal` asks the model to provide a short list of possible commands. If it does, they will be shown.

If you have autocompletions enabled, you can now cycle through the suggestions as autocompletions for the `pal` command:

```
pal # Hit tab now
```

Sometimes a refusal message might be shown if the model can't or won't provide a command suggestion.

### Ask mode

`/ask` mode can be used to pass general queries through to the model, without an expectation that it will suggest shell commands in response.

```
pal /ask Why is the sky blue?
```

### Run mode

By default, `pal` does't have access to the output of other shell commands, such as those it suggests. To feed it the output of some command, use `/run`:

```
pal /run # Command that makes an error
```

Notice how you can add a comment after the command to include a question, instructions, or context.

## Why?

There are other terminal based AI projects, so why build another one? The short answer is that none quite provided the experience I wanted for this particular use case.

One category of in shell assistants drop you into a whole new shell enviroment. This is great for some uses cases, such as the excellent [Aider] project. But I'm still spending a lot of time in my good ol' shell.

Another category leans a bit more toward a TUI, showing a menu of results on the screen to choose from, for example. The experince I desire is closer to a classic shell utility—like an `ls` that lists ideas for the next command to run, instead of listing files.

Finally, who can miss the chance to build something with the incredibly good and cheap DeepSeek API? Times are a changing, so let's have some fun with it :)
