package tests

import (
	"time"
)

type HttpCase struct {
	Name     string            `yaml:"name" json:"name" validate:"required,min=3,max=128"`
	Alias    *string           `yaml:"alias" json:"alias" validate:"omitempty,min=1,max=25"`
	Method   string            `yaml:"method" json:"method" valid:"required,uppercase,oneof=GET POST PUT PATCH DELETE"`
	Endpoint string            `yaml:"endpoint,omitempty" json:"endpoint,omitempty" validate:"omitempty"`
	Url      string            `yaml:"url,omitempty" json:"url,omitempty" validate:"omitempty,url"`
	Headers  map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"  validate:"omitempty,min=1,max=100"`
	Body     map[string]any    `yaml:"body,omitempty" json:"body,omitempty" validate:"omitempty,min=1,max=100"`
	Assert   []*Assert         `yaml:"assert,omitempty" json:"assert,omitempty" validate:"omitempty,min=1,max=50,dive"`
	Save     *Save             `yaml:"save,omitempty" json:"save,omitempty" validate:"omitempty"`
	Timeout  time.Duration     `yaml:"timeout,omitempty" json:"timeout,omitempty" validate:"omitempty,duration"`
	Parallel bool              `yaml:"async,omitempty" json:"async,omitempty" validate:"omitempty,boolean"`
	Details  []string          `yaml:"details,omitempty" json:"details,omitempty" validate:"omitempty,min=1,max=100"`
}

type Assert struct {
	Target   string `yaml:"target,omitempty" json:"target,omitempty" validate:"required,oneof=status body headers"`
	Equals   any    `yaml:"equals,omitempty" json:"equals,omitempty" validate:"omitempty"`
	Contains string `yaml:"contains,omitempty" json:"contains,omitempty" validate:"omitempty,min=1"`
	Exists   bool   `yaml:"exists,omitempty" json:"exists,omitempty" validate:"omitempty,boolean"`
	Template string `yaml:"template,omitempty" json:"template,omitempty" validate:"omitempty,min=1,contains_template"`
}

type Save struct {
	Request  *SaveEntry `yaml:"request,omitempty" json:"request,omitempty" validate:"omitempty"`
	Response *SaveEntry `yaml:"response,omitempty" json:"response,omitempty" validate:"omitempty"`
}

type SaveEntry struct {
	Body    map[string]string `yaml:"body,omitempty" json:"body,omitempty" validate:"omitempty,min=1,max=20,dive,keys,endkeys"`
	Headers []string          `yaml:"headers,omitempty" json:"headers,omitempty"  validate:"omitempty,min=1,max=20"`
}
