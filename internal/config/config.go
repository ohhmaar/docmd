package config

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Version       int              `json:"version"`
	DefaultFolder string           `json:"default_folder_id,omitempty"`
	Links         map[string]*Link `json:"links"`
}

type Link struct {
	DocID           string    `json:"doc_id"`
	DocURL          string    `json:"doc_url"`
	Title           string    `json:"title"`
	CreatedAt       time.Time `json:"created_at"`
	LastSync        time.Time `json:"last_sync"`
	LastRevisionID  string    `json:"last_revision_id,omitempty"`
	LocalHashAtSync string    `json:"local_hash_at_sync,omitempty"`
}

const currentVersion = 1

func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Version: currentVersion,
				Links:   make(map[string]*Link),
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Links == nil {
		cfg.Links = make(map[string]*Link)
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func (c *Config) AddLink(filePath string, link *Link) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	c.Links[absPath] = link
	return c.Save()
}

func (c *Config) GetLink(filePath string) (*Link, bool) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, false
	}

	link, ok := c.Links[absPath]
	return link, ok
}

func (c *Config) RemoveLink(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	delete(c.Links, absPath)
	return c.Save()
}

func (c *Config) UpdateSyncTime(filePath string, revisionID string) error {
	link, ok := c.GetLink(filePath)
	if !ok {
		return nil
	}

	link.LastSync = time.Now()
	link.LastRevisionID = revisionID

	absPath, _ := filepath.Abs(filePath)
	hash, err := HashFile(absPath)
	if err == nil {
		link.LocalHashAtSync = hash
	}

	return c.Save()
}

func HashFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:]), nil
}

func (c *Config) HasLocalChanges(filePath string) (bool, error) {
	link, ok := c.GetLink(filePath)
	if !ok {
		return false, nil
	}

	if link.LocalHashAtSync == "" {
		return true, nil
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false, err
	}

	currentHash, err := HashFile(absPath)
	if err != nil {
		return false, err
	}

	return currentHash != link.LocalHashAtSync, nil
}
