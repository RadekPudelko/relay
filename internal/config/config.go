package config

import (
    "github.com/pelletier/go-toml"
)

// TODO: add config to readme

type Config struct {
    Server ServerConfig `toml:"server"`
    Database DatabaseConfig `toml:"database"`
    Settings SettingsConfig `toml:"settings"`
}

type ServerConfig struct {
    Host string `toml:"address"`
    Port    int    `toml:"port"`
}

type DatabaseConfig struct {
    Filename     string `toml:"filename"`
}

type SettingsConfig struct {
    MaxRoutines       int `toml:"max_routines"`
    PingRetrySeconds  int `toml:"ping_retry_seconds"`
    CFRetrySeconds    int `toml:"cf_retry_seconds"`
    RelayLimit        int `toml:"relay_limit"`
    MaxRetries        int `toml:"max_retries"`
}

func GetDefaultConfig() (Config) {
    return Config{
        Server: ServerConfig{
            Host: "127.0.0.1",
            Port:    8080,
        },
        Database: DatabaseConfig{
            Filename:     "relay.db3",
        },
        Settings: SettingsConfig{
            RelayLimit: 100,
            MaxRoutines: 4,
            PingRetrySeconds: 60,
            CFRetrySeconds: 60,
            MaxRetries: 3,
        },
    }
}

func LoadConfig(filename string, defaultConfig *Config) (*Config, error) {
    config := *defaultConfig

    data, err := toml.LoadFile(filename)
    if err != nil {
        return nil, err
    }

    if err := data.Unmarshal(&config); err != nil {
        return nil, err
    }

    return &config, nil
}

