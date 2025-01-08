package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
)

var (
	Conf Config
)

type Config []Connection

type Connection struct {
	Name       string `json:"name,omitempty"`
	Host       string `json:"host,omitempty"`
	Port       int    `json:"port,omitempty"`
	RemotePath string `json:"remote_path,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
}

func ReadConfig() error {
	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("getting cwd")
		return fmt.Errorf("getting cwd: %w", err)
	}

	file, err := os.Open(path.Join(cwd, ".sync.json"))
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
