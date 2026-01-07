package main

// ModelPricing holds pricing information for different Anthropic models
var ModelPricing = map[string]AnthropicPricing{
	"claude-3-5-haiku-20241022": {
		InputPricePerMTokens:  0.25, // $0.25 per million input tokens
		OutputPricePerMTokens: 1.25, // $1.25 per million output tokens
	},
	"claude-3-haiku-20240307": {
		InputPricePerMTokens:  0.25,
		OutputPricePerMTokens: 1.25,
	},
	"claude-3-5-sonnet-20241022": {
		InputPricePerMTokens:  3.00,
		OutputPricePerMTokens: 15.00,
	},
	"claude-3-opus-20240229": {
		InputPricePerMTokens:  15.00,
		OutputPricePerMTokens: 75.00,
	},
}

// GetPricing returns pricing for a given model name
func GetPricing(modelName string) AnthropicPricing {
	if pricing, ok := ModelPricing[modelName]; ok {
		return pricing
	}
	// Default to Haiku pricing if model not found
	return ModelPricing["claude-3-5-haiku-20241022"]
}

