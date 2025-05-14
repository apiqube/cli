package plan

import (
	"encoding/json"
	"os"
	"time"
)

type StepConfig struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
}

type ExecutionPlan struct {
	Name  string       `json:"name"`
	Steps []StepConfig `json:"steps"`
	Time  time.Time    `json:"time"`
}

func BuildExecutionPlan(_ string) (*ExecutionPlan, error) {
	return &ExecutionPlan{
		Name:  "default-plan",
		Time:  time.Now(),
		Steps: []StepConfig{{Name: "Example", Type: "http", Method: "GET", URL: "http://localhost"}},
	}, nil
}

func SavePlan(plan *ExecutionPlan) error {
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(".testman/plan.json", data, 0644)
}

func LoadPlan() (*ExecutionPlan, error) {
	data, err := os.ReadFile(".apiqube/plan.json")
	if err != nil {
		return nil, err
	}

	var plan ExecutionPlan

	if err = json.Unmarshal(data, &plan); err != nil {
		return nil, err
	}

	return &plan, nil
}
