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
			"claude-3-5-sonnet-latest",
			"claude-3-5-haiku-latest",
			"claude-3-opus-latest",
		},
	},
	"openai": {
		URL: "https://api.openai.com/v1/",
		Models: []string{
			"gpt-4o",
			"chatgpt-4o-latest",
			"gpt-4o-mini",
			"o1",
			"o1-preview",
			"o1-mini",
			"gpt-4",
			"gpt-4-turbo",
			"gpt-4-turbo-preview",
		},
	},
	"mistral": {
		URL: "https://api.mistral.ai/v1/",
		Models: []string{
			"codestral-latest",
			"mistral-large-latest",
			"open-mistral-nemo",
			"mistral-small-latest",
			"open-codestral-mamba",
		},
	},
	"google": {
		URL: "https://generativelanguage.googleapis.com/v1beta/openai/",
		Models: []string{
			"gemini-2.0-pro-exp-02-05",
			"gemini-2.0-flash-thinking-exp-01-21",
			"gemini-2.0-flash-001",
			"gemini-2.0-flash-lite-preview-02-05",
			"gemini-1.5-flash",
			"gemini-1.5-flash-8b",
			"gemini-1.5-pro",
		},
	},
}
