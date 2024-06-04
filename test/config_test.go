package test

import (
	"testing"
    "os"

	"relay/internal/config"
)

func TestConfig(t *testing.T) {
	t.Log("TestConfig")

    // Write the partial configuration to a temporary file
    tmpfile, err := os.CreateTemp("", "test_config.toml")
    if err != nil {
        t.Fatalf("TestConfig: os.CreateTemp: %v", err)
    }
    defer os.Remove(tmpfile.Name()) // Clean up the temporary file afterwards

    partialConfig := getConfigString()
    if _, err := tmpfile.Write([]byte(partialConfig)); err != nil {
        t.Fatalf("TestConfig: tmpfile.Write: %v", err)
    }
    if err := tmpfile.Close(); err != nil {
        t.Fatalf("TestConfig: tmpfile.Close(): %v", err)
    }

    defaultConfig := config.GetDefaultConfig()
    myConfig, err := config.LoadConfig(tmpfile.Name(), &defaultConfig)
    if err != nil {
        t.Fatalf("TestConfig: %v", err)
    }

    if myConfig.Server.Host != defaultConfig.Server.Host {
        t.Errorf("TestConfig: server address, want=%s, got=%s", defaultConfig.Server.Host, myConfig.Server.Host)
    }
    if myConfig.Server.Port != 8081 {
        t.Errorf("TestConfig: server port, want=8081, got=%d", myConfig.Server.Port)
    }
    if myConfig.Database.Filename != "my.db3" {
        t.Errorf("TestConfig: database filename, want=my.db3, got=%s", myConfig.Database.Filename)
    }
    if myConfig.Settings.MaxRoutines != defaultConfig.Settings.MaxRoutines {
        t.Errorf("TestConfig: settings MaxRoutines, want=%d, got=%d", defaultConfig.Settings.MaxRoutines, myConfig.Settings.PingRetrySeconds)
    }
    if myConfig.Settings.PingRetrySeconds != 5 {
        t.Errorf("TestConfig: settings PingRetrySeconds, want=5, got=%d", myConfig.Settings.PingRetrySeconds)
    }
    if myConfig.Settings.CFRetrySeconds != defaultConfig.Settings.CFRetrySeconds {
        t.Errorf("TestConfig: settings CFRetrySeconds, want=%d, got=%d", defaultConfig.Settings.CFRetrySeconds, myConfig.Settings.CFRetrySeconds)
    }
    if myConfig.Settings.MaxRetries != defaultConfig.Settings.MaxRetries {
        t.Errorf("TestConfig: settings MaxRetries, want=%d, got=%d", defaultConfig.Settings.MaxRetries, myConfig.Settings.MaxRetries)
    }
}

func getConfigString() (string) {
    return `
[server]
port = 8081

[database]
filename = "my.db3"

[settings]
ping_retry_seconds = 5
`
}
