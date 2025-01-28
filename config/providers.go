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
		URL: "https://api.deepseek.ai/v1",
		Models: []string{
			"deepseek-chat",
		},
	},
	"openai": {
		URL: "https://api.openai.com/v1",
		Models: []string{
			"gpt-3.5-turbo",
			"gpt-4",
			"gpt-4-turbo-preview",
		},
	},
	"anthropic": {
		URL: "https://api.anthropic.com/v1",
		Models: []string{
			"claude-2",
			"claude-instant-1",
			"claude-3-opus",
			"claude-3-sonnet",
		},
	},
	"huggingface": {
		URL: "https://api-inference.huggingface.co/v1/",
		Models: []string{
			"meta-llama/Llama-3.3-70B-Instruct",
			"meta-llama/Llama-3.1-8B-Instruct",
			"meta-llama/Llama-3.2-3B-Instruct",
			"meta-llama/Llama-2-7b-chat-hf",
			"deepseek-ai/DeepSeek-R1-Distill-Qwen-32B",
			"deepseek-ai/DeepSeek-R1-Distill-Qwen-1.5B",
		},
	},
}
