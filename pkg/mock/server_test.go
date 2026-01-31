package mock_test

import (
    "io"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/nexusapi/nexus/pkg/mock"
)

func TestMockServer_ServeHTTP(t *testing.T) {
    srv := mock.NewServer()

    srv.AddEndpoint(&mock.Endpoint{
        Path:   "/health",
        Method: "GET",
        Response: mock.Response{
            StatusCode: 200,
            Body:       "OK",
        },
    })

    ts := httptest.NewServer(srv)
    defer ts.Close()

    res, err := http.Get(ts.URL + "/health")
    if err != nil {
        t.Fatalf("http get: %v", err)
    }
    defer res.Body.Close()

    if res.StatusCode != 200 {
        t.Fatalf("expected 200 got %d", res.StatusCode)
    }

    b, _ := io.ReadAll(res.Body)
    if string(b) != "OK" {
        t.Fatalf("unexpected body: %s", string(b))
    }
}
