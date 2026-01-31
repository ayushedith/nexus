package http_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    nexushttp "github.com/nexusapi/nexus/pkg/http"
)

func TestClientDo_GET(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/plain")
        w.WriteHeader(200)
        w.Write([]byte("ok"))
    }))
    defer ts.Close()

    client := nexushttp.NewClient(nil)

    ctx := context.Background()
    resp, err := client.Do(ctx, &nexushttp.RequestOptions{
        Method: "GET",
        URL:    ts.URL,
    })
    if err != nil {
        t.Fatalf("Do() error: %v", err)
    }

    if resp.StatusCode != 200 {
        t.Fatalf("expected 200 got %d", resp.StatusCode)
    }
    if string(resp.Body) != "ok" {
        t.Fatalf("unexpected body: %s", string(resp.Body))
    }
}
