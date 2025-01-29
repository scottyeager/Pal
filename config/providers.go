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
		URL: "https://api.deepseek.com",
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
		URL: "https://api.anthropic.com/v1",
		Models: []string{
			"claude-3-opus-20240229",
			"claude-3-sonnet-20240229",
			"claude-2.1",
			"claude-2.0",
			"claude-instant-1.2",
		},
	},
}
