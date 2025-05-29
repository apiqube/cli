package tests

import (
	"time"
)

type HttpCase struct {
	Name     string            `yaml:"name" json:"name" validate:"required,min=3,max=128"`
	Method   string            `yaml:"method" json:"method" valid:"required,uppercase,oneof=GET POST PUT PATCH DELETE"`
	Endpoint string            `yaml:"endpoint,omitempty" json:"endpoint,omitempty" validate:"omitempty"`
	Url      string            `yaml:"url,omitempty" json:"url,omitempty" validate:"omitempty,url"`
	Headers  map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"  validate:"omitempty,min=1,max=100"`
	Body     map[string]any    `yaml:"body,omitempty" json:"body,omitempty" validate:"omitempty,min=1,max=100"`
	Assert   []*Assert         `yaml:"assert,omitempty" json:"assert,omitempty" validate:"omitempty,min=1,max=50,dive"`
	Save     *Save             `yaml:"save,omitempty" json:"save,omitempty" validate:"omitempty"`
	Pass     []Pass            `yaml:"pass,omitempty" json:"pass,omitempty" validate:"omitempty,min=1,max=25,dive"`
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
	Json    map[string]string `yaml:"json,omitempty" json:"json,omitempty" validate:"omitempty,dive,keys,endkeys"`
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"  validate:"omitempty,dive,keys,endkeys"`
	Status  bool              `yaml:"status,omitempty" json:"status,omitempty" validate:"omitempty,boolean"`
	Body    bool              `yaml:"body,omitempty" json:"body,omitempty" validate:"omitempty,boolean"`
	All     bool              `yaml:"all,omitempty" json:"all,omitempty" validate:"omitempty,boolean"`
	Group   string            `yaml:"group,omitempty" json:"group,omitempty" validate:"omitempty,min=1,max=100"`
}

type Pass struct {
	From string            `yaml:"from" json:"from" validate:"required,min=1,max=100"`
	Map  map[string]string `yaml:"map,omitempty" json:"map,omitempty" validate:"omitempty,min=1,max=100,dive,keys,endkeys"`
}
