package config

import (
	"github.com/CVDS2020/CVDS2020/common/config"
	"github.com/CVDS2020/CVDS2020/common/def"
)

type Service struct {
	Name        string `yaml:"name" json:"name"`
	DisplayName string `yaml:"display-name" json:"display-name"`
	Description string `yaml:"description" json:"description"`
}

func (s *Service) PreHandle() config.PreHandlerConfig {
	if s == nil {
		s = new(Service)
	}
	s.Name = "CVDSMDU"
	s.Description = "Media Distribute Unit Application"
	return s
}

func (s *Service) PostHandle() (config.PostHandlerConfig, error) {
	def.SetDefault(&s.DisplayName, s.Name)
	return s, nil
}
