package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nexusapi/nexus/pkg/ai"
	"github.com/nexusapi/nexus/pkg/collab"
	"github.com/nexusapi/nexus/pkg/collection"
	"github.com/nexusapi/nexus/pkg/mock"
	"github.com/nexusapi/nexus/pkg/storage"
	"github.com/nexusapi/nexus/pkg/tui"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "tui":
		runTUI()
	case "run":
		runCLI()
	case "load":
		runLoadTest()
	case "mock":
		runMockServer()
	case "collab":
		runCollabServer()
	case "ai":
		runAI()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: nexus <command> [args]")
	fmt.Println("\nCommands:")
	fmt.Println("  tui <collection>              - Start terminal UI")
	fmt.Println("  run <collection>              - Run collection from CLI")
	fmt.Println("  load <collection>             - Run load test")
	fmt.Println("  mock                          - Start mock server")
	fmt.Println("  collab                        - Start collaboration server")
	fmt.Println("\nAI Commands:")
	fmt.Println("  ai generate-body <schema>     - Generate request body from schema")
	fmt.Println("  ai generate-tests <spec>      - Generate test assertions")
	fmt.Println("  ai optimize <request-info>    - Suggest request optimizations")
	fmt.Println("  ai from-description <desc>    - Convert natural language to collection")
	fmt.Println("  ai analyze-changes <old> <new> - Analyze API changes")
	fmt.Println("\nOptions:")
	fmt.Println("  --env <environment>           - Environment to use (default: dev)")
	fmt.Println("  --api-key <key>               - OpenAI API key (or use OPENAI_API_KEY env)")
}

func runTUI() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: nexus tui <collection>")
		os.Exit(1)
	}

	collectionPath := os.Args[2]
	env := getEnv()

	repo, err := storage.NewRepository(".")
	if err != nil {
		log.Fatal(err)
	}

	coll, err := repo.LoadCollection(collectionPath)
	if err != nil {
		log.Fatal(err)
	}

	model := tui.NewModel(coll, env)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func runCLI() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: nexus run <collection>")
		os.Exit(1)
	}

	collectionPath := os.Args[2]
	env := getEnv()

	parser := collection.NewParser()
	coll, err := parser.ParseFile(collectionPath)
	if err != nil {
		log.Fatal(err)
	}

	runner := collection.NewRunner(env)
	results, err := runner.Run(coll)
	if err != nil {
		log.Fatal(err)
	}

	passed := 0
	failed := 0

	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("❌ %s: %v\n", result.Request.Name, result.Error)
			failed++
			continue
		}

		if result.Passed {
			fmt.Printf("✅ %s: %s (%v)\n", result.Request.Name, result.Response.Status, result.Response.Time)
			passed++
		} else {
			fmt.Printf("⚠️  %s: %s (%v) - Assertions failed: %v\n",
				result.Request.Name, result.Response.Status, result.Response.Time, result.Failures)
			failed++
		}
	}

	fmt.Printf("\nResults: %d passed, %d failed\n", passed, failed)

	if failed > 0 {
		os.Exit(1)
	}
}

func runLoadTest() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: nexus load <collection> [flags]")
		os.Exit(1)
	}

	fmt.Println("Load testing: Basic implementation")

	collectionPath := os.Args[2]
	env := getEnv()

	parser := collection.NewParser()
	coll, err := parser.ParseFile(collectionPath)
	if err != nil {
		log.Fatal(err)
	}

	runner := collection.NewRunner(env)

	fmt.Println("Running 10 iterations...")

	for i := 0; i < 10; i++ {
		results, err := runner.Run(coll)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Iteration %d: %d requests completed\n", i+1, len(results))
	}

	fmt.Println("Load test complete!")
}

func runMockServer() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: nexus mock [port]")
		os.Exit(1)
	}

	port := "9999"
	if len(os.Args) > 2 {
		port = os.Args[2]
	}

	fmt.Printf("Starting mock server on port %s...\n", port)

	server := mock.NewServer()

	server.AddEndpoint(&mock.Endpoint{
		Path:   "/health",
		Method: "GET",
		Response: mock.Response{
			StatusCode: 200,
			Body:       "OK",
		},
	})

	server.AddEndpoint(&mock.Endpoint{
		Path:   "/api/users",
		Method: "GET",
		Response: mock.Response{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: map[string]interface{}{
				"users": []map[string]interface{}{
					{"id": 1, "name": "Alice", "email": "alice@example.com"},
					{"id": 2, "name": "Bob", "email": "bob@example.com"},
				},
			},
		},
	})

	addr := ":" + port
	fmt.Printf("Mock server running at http://localhost%s\n", addr)
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /health")
	fmt.Println("  GET  /api/users")

	if err := server.Start(addr); err != nil {
		log.Fatal(err)
	}
}

func runCollabServer() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: nexus collab [port]")
		os.Exit(1)
	}

	port := "8080"
	if len(os.Args) > 2 {
		port = os.Args[2]
	}

	fmt.Printf("Starting collaboration server on port %s...\n", port)

	server := collab.NewServer()

	http.HandleFunc("/ws", server.HandleWebSocket)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := ":" + port
	fmt.Printf("WebSocket server running at ws://localhost%s/ws\n", addr)
	fmt.Println("Connect with: ?room=<room-id>")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func getEnv() string {
	env := os.Getenv("NEXUS_ENV")
	if env == "" {
		env = "dev"
	}
	return env
}

func runAI() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: nexus ai <command> [args]")
		fmt.Println("\nAI Commands:")
		fmt.Println("  generate-body <schema>      - Generate request body from schema")
		fmt.Println("  generate-tests <spec>       - Generate test assertions")
		fmt.Println("  optimize <request-info>     - Suggest request optimizations")
		fmt.Println("  from-description <desc>     - Convert natural language to collection")
		fmt.Println("  analyze-changes <old> <new> - Analyze API changes")
		os.Exit(1)
	}

	subcommand := os.Args[2]

	apiKey := os.Getenv("OPENAI_API_KEY")
	for i, arg := range os.Args {
		if arg == "--api-key" && i+1 < len(os.Args) {
			apiKey = os.Args[i+1]
			break
		}
	}

	client := ai.NewOpenAIClient(apiKey)

	var aiClient ai.AIClient = client

	if !aiClient.IsConfigured() {
		fmt.Println(aiClient.GetSetupInstructions())
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	switch subcommand {
	case "generate-body":
		aiGenerateBody(ctx, client)
	case "generate-tests":
		aiGenerateTests(ctx, client)
	case "optimize":
		aiOptimize(ctx, client)
	case "from-description":
		aiFromDescription(ctx, client)
	case "analyze-changes":
		aiAnalyzeChanges(ctx, client)
	default:
		fmt.Printf("Unknown AI command: %s\n", subcommand)
		os.Exit(1)
	}
}

func aiGenerateBody(ctx context.Context, client ai.AIClient) {
	if len(os.Args) < 4 {
		fmt.Println("Usage: nexus ai generate-body <schema>")
		os.Exit(1)
	}

	schema := os.Args[3]
	fmt.Println("Generating request body...")

	body, err := client.GenerateRequestBody(ctx, schema)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nGenerated body:")
	fmt.Println(body)
}

func aiGenerateTests(ctx context.Context, client ai.AIClient) {
	if len(os.Args) < 4 {
		fmt.Println("Usage: nexus ai generate-tests <spec>")
		os.Exit(1)
	}

	spec := os.Args[3]
	fmt.Println("Generating test assertions...")

	tests, err := client.GenerateTests(ctx, spec)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nGenerated tests:")
	for _, test := range tests {
		fmt.Printf("  - %s\n", test)
	}
}

func aiOptimize(ctx context.Context, client ai.AIClient) {
	if len(os.Args) < 4 {
		fmt.Println("Usage: nexus ai optimize <request-info>")
		os.Exit(1)
	}

	requestInfo := os.Args[3]
	fmt.Println("Analyzing request for optimizations...")

	suggestions, err := client.SuggestOptimizations(ctx, requestInfo)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nOptimization suggestions:")
	for _, suggestion := range suggestions {
		fmt.Printf("  - %s\n", suggestion)
	}
}

func aiFromDescription(ctx context.Context, client ai.AIClient) {
	if len(os.Args) < 4 {
		fmt.Println("Usage: nexus ai from-description <description>")
		os.Exit(1)
	}

	description := os.Args[3]
	fmt.Println("Generating collection from description...")

	collection, err := client.GenerateFromNaturalLanguage(ctx, description)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nGenerated collection:")
	fmt.Println(collection)
}

func aiAnalyzeChanges(ctx context.Context, client ai.AIClient) {
	if len(os.Args) < 5 {
		fmt.Println("Usage: nexus ai analyze-changes <old-spec> <new-spec>")
		os.Exit(1)
	}

	oldSpec := os.Args[3]
	newSpec := os.Args[4]
	fmt.Println("Analyzing API changes...")

	analysis, err := client.AnalyzeAPIChanges(ctx, oldSpec, newSpec)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nAPI Change Analysis:")
	fmt.Println(analysis)
}
