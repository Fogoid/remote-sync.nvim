package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
)

var (
	Conf Config
)

type Config struct {
	Connections []Connection `json:"connections"`
}

type Connection struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	RemotePath string `json:"remote_path"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

func ReadConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		slog.Error("opening file")
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	configBytes, err := io.ReadAll(file)
	if err != nil {
		slog.Error("reading file")
		return fmt.Errorf("reading file: %w", err)
	}

	err = json.Unmarshal(configBytes, &Conf)
	if err != nil {
		slog.Error("unmarshling")
		return fmt.Errorf("unmarshling: %w", err)
	}

	return nil
}
