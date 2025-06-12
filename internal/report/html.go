package report

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/apiqube/cli/internal/core/runner/save"
)

//go:embed html/templates/*.gohtml
var htmlTemplates embed.FS

// CaseReport is a view model for a single test case.
type CaseReport struct {
	Name       string
	Success    bool
	Assert     string
	StatusCode int
	Duration   time.Duration
	Errors     []string
	Method     string
	Request    *save.Entry
	Response   *save.Entry
}

// ManifestReport groups results for a single manifest.
type ManifestReport struct {
	ManifestID  string
	Target      string
	TotalCases  int
	PassedCases int
	FailedCases int
	TotalTime   time.Duration
	Cases       []*CaseReport
}

// ViewData is the data passed to the HTML template.
type ViewData struct {
	GeneratedAt   time.Time
	TotalCases    int
	PassedCases   int
	FailedCases   int
	TotalTime     time.Duration
	ManifestStats []*ManifestReport
}

// BuildReportViewData aggregates results and statistics for the template.
func BuildReportViewData(results []*save.Result) *ViewData {
	manifestMap := make(map[string]*ManifestReport)
	totalCases := 0
	passedCases := 0
	failedCases := 0
	totalTime := time.Duration(0)

	for _, res := range results {
		m, ok := manifestMap[res.ManifestID]
		if !ok {
			m = &ManifestReport{
				ManifestID: res.ManifestID,
				Target:     res.Target,
				Cases:      []*CaseReport{},
			}
			manifestMap[res.ManifestID] = m
		}
		cr := res.ResultCase
		caseReport := &CaseReport{
			Name:       cr.Name,
			Success:    cr.Success,
			Assert:     cr.Assert,
			StatusCode: cr.StatusCode,
			Duration:   cr.Duration,
			Errors:     cr.Errors,
			Method:     res.Method,
			Request:    res.Request,
			Response:   res.Response,
		}
		m.Cases = append(m.Cases, caseReport)
		m.TotalCases++
		m.TotalTime += cr.Duration
		if cr.Success {
			m.PassedCases++
			passedCases++
		} else {
			m.FailedCases++
			failedCases++
		}
		totalCases++
		totalTime += cr.Duration
	}
	manifestStats := make([]*ManifestReport, 0, len(manifestMap))
	for _, m := range manifestMap {
		manifestStats = append(manifestStats, m)
	}
	return &ViewData{
		GeneratedAt:   time.Now(),
		TotalCases:    totalCases,
		PassedCases:   passedCases,
		FailedCases:   failedCases,
		TotalTime:     totalTime,
		ManifestStats: manifestStats,
	}
}

// HTMLReportGenerator generates an HTML report using html/template and gohtml templates.
type HTMLReportGenerator struct {
	tmpl  *template.Template
	funcs template.FuncMap
}

// NewHTMLReportGenerator creates a new HTMLReportGenerator with custom functions and template directory.
func NewHTMLReportGenerator() (*HTMLReportGenerator, error) {
	funcs := template.FuncMap{
		"formatTime": func(t time.Time) string { return t.Format("2006-01-02 15:04:05") },
		"statusText": func(success bool) string {
			if success {
				return "PASSED"
			}
			return "FAILED"
		},
		// Add more custom functions here as needed
	}

	tmpl, err := template.New("base.gohtml").Funcs(funcs).ParseFS(htmlTemplates, "html/templates/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &HTMLReportGenerator{
		tmpl:  tmpl,
		funcs: funcs,
	}, nil
}

// Generate creates an HTML report from test results and writes it to outputPath.
func (g *HTMLReportGenerator) Generate(results []*save.Result) error {
	data := BuildReportViewData(results)

	outputPath := filepath.Join("C:\\Users\\admin\\Desktop\\reports", fmt.Sprintf("/report_%s.html", time.Now().Format("2006-01-02-150405")))
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}

	defer func() {
		_ = file.Close()
	}()

	if err = g.tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
