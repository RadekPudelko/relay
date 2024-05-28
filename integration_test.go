package main

import (
	"testing"
    "net/http"
    "io"

	// "pcfs/db"
	// "pcfs/particle"
	// "pcfs/server"
)

func TestIntegration(t *testing.T) {
    t.Log("TestIntegration")
    go func() {
        err := runTestServer()
        if err != nil {
            // TODO: Fix this warning
            t.Fatalf("TestIntegration: %+v", err)
        }
    }()

    res, err := http.Get("http://localhost:8080/")
    if err != nil {
        t.Fatalf("TestIntegration: http.Get: %+v", err)
    }
    defer res.Body.Close()

    out, err := io.ReadAll(res.Body)
    if err != nil {
        t.Fatalf("TestIntegration: io.ReadAll: %+v", err)
    }
    t.Logf("TestIntegration: response from server: %s\n", string(out))

}

