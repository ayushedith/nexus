package api

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/nexusapi/nexus/pkg/ai"
    "github.com/nexusapi/nexus/pkg/collection"
    "github.com/nexusapi/nexus/pkg/log"
    "github.com/nexusapi/nexus/pkg/metrics"
    "github.com/nexusapi/nexus/pkg/mock"
    "github.com/nexusapi/nexus/pkg/storage"
)

type APIServer struct {
    repo       *storage.Repository
    aiClient   ai.AIClient
    mockServer *mock.Server
    env        string
}

func NewAPIServer(basePath, env, apiKey string) (*APIServer, error) {
    log.Init()

    repo, err := storage.NewRepository(basePath)
    if err != nil {
        return nil, fmt.Errorf("setup storage: %w", err)
    }

    // local client adapter available in pkg/ai
    var aiClient ai.AIClient = ai.NewClient("")
    if apiKey != "" {
        aiClient = ai.NewOpenAIClient(apiKey)
    }

    return &APIServer{
        repo:       repo,
        aiClient:   aiClient,
        mockServer: mock.NewServer(),
        env:        env,
    }, nil
}

// RegisterRoutes registers HTTP handlers on the provided mux.
func (s *APIServer) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/api/collections", s.handleCollections)
    mux.HandleFunc("/api/collections/get", s.handleGetCollection)
    mux.HandleFunc("/api/collections/save", s.handleSaveCollection)
    mux.HandleFunc("/api/run", s.handleRun)
    mux.HandleFunc("/api/mock/add", s.handleMockAdd)
    mux.HandleFunc("/api/ai/generate-body", s.handleAIGenerateBody)
    mux.Handle("/metrics", metrics.Handler())
}

func writeJSON(w http.ResponseWriter, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(v)
}

func (s *APIServer) handleCollections(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        cols, err := s.repo.ListCollections()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        writeJSON(w, map[string]interface{}{"collections": cols})
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}

func (s *APIServer) handleGetCollection(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("name")
    if name == "" {
        http.Error(w, "missing name", http.StatusBadRequest)
        return
    }
    coll, err := s.repo.LoadCollection(name)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    writeJSON(w, coll)
}

type saveReq struct {
    Name    string `json:"name"`
    Content string `json:"content"`
}

func (s *APIServer) handleSaveCollection(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var sr saveReq
    if err := json.NewDecoder(r.Body).Decode(&sr); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    coll, err := collection.NewParser().ParseBytes([]byte(sr.Content))
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := s.repo.SaveCollection(coll, sr.Name); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    writeJSON(w, map[string]string{"status": "ok"})
}

type runReq struct {
    Name    string `json:"name,omitempty"`
    Content string `json:"content,omitempty"`
    Env     string `json:"env,omitempty"`
}

func (s *APIServer) handleRun(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var rr runReq
    if err := json.NewDecoder(r.Body).Decode(&rr); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    var coll *collection.Collection
    var err error
    parser := collection.NewParser()
    if rr.Name != "" {
        coll, err = s.repo.LoadCollection(rr.Name)
        if err != nil {
            http.Error(w, err.Error(), http.StatusNotFound)
            return
        }
    } else if rr.Content != "" {
        coll, err = parser.ParseBytes([]byte(rr.Content))
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
    } else {
        http.Error(w, "missing name or content", http.StatusBadRequest)
        return
    }

    env := s.env
    if rr.Env != "" {
        env = rr.Env
    }

    runner := collection.NewRunner(env)
    results, err := runner.Run(coll)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // update metrics
    for _, res := range results {
        status := "ok"
        if res.Error != nil || !res.Passed {
            status = "failed"
        }
        metrics.RequestsTotal.WithLabelValues(status).Inc()
    }

    writeJSON(w, map[string]interface{}{"results": results})
}

func (s *APIServer) handleMockAdd(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var ep mock.Endpoint
    if err := json.NewDecoder(r.Body).Decode(&ep); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    s.mockServer.AddEndpoint(&ep)
    writeJSON(w, map[string]string{"status": "ok"})
}

type aiReq struct {
    Schema string `json:"schema"`
}

func (s *APIServer) handleAIGenerateBody(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var ar aiReq
    if err := json.NewDecoder(r.Body).Decode(&ar); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    body, err := s.aiClient.GenerateRequestBody(ctx, ar.Schema)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    writeJSON(w, map[string]string{"body": body})
}
