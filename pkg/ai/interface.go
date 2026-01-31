package ai

import "context"

// AIClient defines the methods an AI adapter must implement.
type AIClient interface {
    IsConfigured() bool
    GetSetupInstructions() string

    CreateMessage(ctx context.Context, prompt string, maxTokens int) (string, error)
    GenerateRequestBody(ctx context.Context, schema string) (string, error)
    GenerateTests(ctx context.Context, apiSpec string) ([]string, error)
    SuggestOptimizations(ctx context.Context, requestInfo string) ([]string, error)
    GenerateFromNaturalLanguage(ctx context.Context, description string) (string, error)
    AnalyzeAPIChanges(ctx context.Context, oldSpec, newSpec string) (string, error)
}
