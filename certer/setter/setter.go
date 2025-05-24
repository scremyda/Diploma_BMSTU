package setter

import (
	"context"
	"diploma/models"
	"fmt"
	"os"
	"path/filepath"
)

type Interface interface {
	Set(ctx context.Context, domain, cert, key string) error
}

type Settings struct {
	Path string `yaml:"path"`
	Type string `yaml:"type"`
}

type Config struct {
	Sets map[models.Domain]Settings `yaml:"sets"`
}

type Setter struct {
	conf Config
}

func New(conf Config) *Setter {
	return &Setter{
		conf: conf,
	}
}

func (s *Setter) Set(ctx context.Context, domain, cert, key string) error {
	settings, ok := s.conf.Sets[models.Domain(domain)]
	if !ok {
		return fmt.Errorf("no setter configuration found for domain %s", domain)
	}

	if err := os.MkdirAll(settings.Path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", settings.Path, err)
	}

	certFilePath := filepath.Join(settings.Path, "localhost.crt") //"cert.pem"
	keyFilePath := filepath.Join(settings.Path, "localhost.key")  //"key.pem"

	// Записываем сертификат в файл.
	if err := os.WriteFile(certFilePath, []byte(cert), 0644); err != nil {
		return fmt.Errorf("failed to write certificate file: %w", err)
	}

	// Записываем приватный ключ в файл.
	if err := os.WriteFile(keyFilePath, []byte(key), 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}
