package config

type Provider struct {
	URL    string   `yaml:"url"`
	APIKey string   `yaml:"api_key"`
	Models []string `yaml:"models"`
}

func NewProvider(providerName string, apiKey string) Provider {
	template := ProviderTemplates[providerName]
	template.APIKey = apiKey
	return template
}

var ProviderTemplates = map[string]Provider{
	"deepseek": {
		URL: "https://api.deepseek.com/",
		Models: []string{
			"deepseek-chat",
			"deepseek-reasoner",
		},
	},
	"huggingface": {
		URL: "https://api-inference.huggingface.co/v1/",
		Models: []string{
			"meta-llama/Llama-3.3-70B-Instruct",
			"meta-llama/Llama-3.2-3B-Instruct",
			"meta-llama/Llama-2-7b-chat-hf",
			"deepseek-ai/DeepSeek-R1-Distill-Qwen-32B",
			"deepseek-ai/DeepSeek-R1-Distill-Qwen-1.5B",
		},
	},
	"anthropic": {
		// Anthropic SDK requires no trailing slash, while OpenAI needs it
		// We might want to let it connect automatically since it's using it's
		// native SDK
		URL: "https://api.anthropic.com/v1",
		Models: []string{
			"claude-opus-4-0",
			"claude-sonnet-4-0",
			"claude-3-7-sonnet-latest",
			"claude-3-5-sonnet-latest",
			"claude-3-5-haiku-latest",
		},
	},
	"openai": {
		URL: "https://api.openai.com/v1/",
		Models: []string{
			"gpt-4.1",
			"gpt-4.1-mini",
			"gpt-4.1-nano",
			"gpt-4o",
			"gpt-4o-mini",
			"chatgpt-4o-latest",
			"o4-mini",
			"o3",
			"o3-pro",
			"o3-mini",
		},
	},
	"mistral": {
		URL: "https://api.mistral.ai/v1/",
		Models: []string{
			"magistral-medium-latest",
			"magistral-small-latest",
			"mistral-medium-latest",
			"mistral-large-latest",
			"mistral-small-latest",
			"devstral-small-latest",
			"devstral-medium-latest",
			"codestral-latest",
			"mistral-large-latest",
			"open-mistral-nemo",
			"mistral-small-latest",
		},
	},
	"google": {
		URL: "https://generativelanguage.googleapis.com/v1beta/openai/",
		Models: []string{
			"gemini-2.5-pro",
			"gemini-2.5-flash",
			"gemini-2.5-flash-lite-preview-06-17",
			"gemini-2.0-flash",
			"gemini-2.0-flash-lite",
			"gemini-1.5-flash",
			"gemini-1.5-flash-8b",
			"gemini-1.5-pro",
		},
	},
}
