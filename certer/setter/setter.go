package setter

import (
	"context"
	"diploma/models"
)

type Interface interface {
	Set(ctx context.Context, cert, key string) error
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
