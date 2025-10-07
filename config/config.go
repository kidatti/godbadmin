package config

import (
	"encoding/json"
	"os"
	"sync"
)

type ServerConfig struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	DBType   string `json:"db_type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type Settings struct {
	Servers       []ServerConfig `json:"servers"`
	EncryptionKey string         `json:"encryption_key,omitempty"`
	mu            sync.RWMutex
}

var (
	settings     *Settings
	settingsOnce sync.Once
)

func GetSettings() *Settings {
	settingsOnce.Do(func() {
		settings = &Settings{
			Servers: []ServerConfig{},
		}
	})
	return settings
}

func (s *Settings) Load(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, s); err != nil {
		return err
	}

	// Decrypt passwords
	for i := range s.Servers {
		if s.Servers[i].Password != "" {
			decrypted, err := Decrypt(s.Servers[i].Password)
			if err != nil {
				// If decryption fails, assume it's plain text (for backward compatibility)
				// and encrypt it on next save
				continue
			}
			s.Servers[i].Password = decrypted
		}
	}

	return nil
}

func (s *Settings) Save(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a copy with encrypted passwords
	encrypted := Settings{
		Servers:       make([]ServerConfig, len(s.Servers)),
		EncryptionKey: s.EncryptionKey, // Preserve encryption key
	}
	copy(encrypted.Servers, s.Servers)

	// Encrypt passwords
	for i := range encrypted.Servers {
		if encrypted.Servers[i].Password != "" {
			encryptedPwd, err := Encrypt(encrypted.Servers[i].Password)
			if err != nil {
				return err
			}
			encrypted.Servers[i].Password = encryptedPwd
		}
	}

	data, err := json.MarshalIndent(encrypted, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func (s *Settings) AddServer(server ServerConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Servers = append(s.Servers, server)
}

func (s *Settings) UpdateServer(id string, server ServerConfig) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, srv := range s.Servers {
		if srv.ID == id {
			s.Servers[i] = server
			return true
		}
	}
	return false
}

func (s *Settings) DeleteServer(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, srv := range s.Servers {
		if srv.ID == id {
			s.Servers = append(s.Servers[:i], s.Servers[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Settings) GetServer(id string) (*ServerConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, srv := range s.Servers {
		if srv.ID == id {
			return &srv, true
		}
	}
	return nil, false
}

func (s *Settings) GetServers() []ServerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	servers := make([]ServerConfig, len(s.Servers))
	copy(servers, s.Servers)
	return servers
}
