package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shotah/go-google-home-mcp/home"
)

func sessionPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("HOME")
	}
	return filepath.Join(configDir, "ghome", "session.json")
}

func loadClient() (*home.Client, error) {
	path := sessionPath()
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("not logged in, run: ghome login")
	}
	defer f.Close()

	client := home.New(home.Options{})
	if err := client.LoadSession(f); err != nil {
		return nil, fmt.Errorf("session corrupted: %w", err)
	}
	client.SetSessionPersister(saveClient)
	return client, nil
}

func saveClient(client *home.Client) error {
	path := sessionPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	return client.SaveSession(f)
}

func removeSession() error {
	return os.Remove(sessionPath())
}

func jsonEncoder() *json.Encoder {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc
}
