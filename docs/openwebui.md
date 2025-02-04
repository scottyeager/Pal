# Connect OpenWebUI to Pal

The [OpenWebUI](https://openwebui.com/) project is a popular front end for self hosted LLMs. It has a variety of other features, but here we'll see how to use OpenWebUI to connect models running in Ollama to `pal`.

## Prerequisites

This guide assumes you already have OpenWebUI installed and setup with at least one model. If you didn't get that far yet, follow along the [OpenWebUI docs](https://docs.openwebui.com/) until that's done and then pick back up here.

## Get OpenWebUI API key

Like commercial LLM providers, OpenWebUI also provides an API key that you can use to access your instance. This can be found in the menu under *Settings > Account*:

![image](https://github.com/user-attachments/assets/5528c50a-8f91-45f5-bbe4-13c6b3072582)

Keep this page open. We'll need it in a sec.

## Add pal config file entry

Next we'll manually add an entry for OpenWebUI in the `pal` config file. Edit the config file:

```
$EDITOR ~/.config/pal_helper/config.yaml
```

Add a section like this to the end of the file:

```yaml
    openwebui:
        url: http://127.0.0.1:3000/api/
        api_key: sk-abc123
        models:
            - deepseek-r1:1.5b
            - deepseek-r1:7b
```

If OpenWebUI is running on a different machine or is using a different port, adjust the url accordingly. Be sure to include the slash at the end: `api/`. Copy and paste your API key from earlier onto the `api_key` line.

Enter as many models as you want. These models must already be installed and the names must match exactly. You can easily copy and paste the model name from the new chat page:

![image](https://github.com/user-attachments/assets/fa4d696f-7cff-4af5-b447-810f7aed4304)

## Select the model

After saving the config file, you should now be able to choose the configured OpenWebUI models using `pal /models`. Depending on your hardware specs, the model you are using, and the length of the prompt and response, it can sometimes take a while to get a reply.

With some brief testing, I found that even the little `deepseek-r1:1.5b` (Qwen 1.5B fine tuned by DeepSeek R1) can provide coherent command suggestions. Nice 😎
