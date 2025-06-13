package html

import (
	"embed"
	"fmt"
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"github.com/goccy/go-json"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/apiqube/cli/internal/core/runner/save"
)

//go:embed templates/*.gohtml
var htmlTemplates embed.FS

// CaseReport is a view model for a single test case.
type CaseReport struct {
	Name       string
	Method     string
	Success    bool
	Assert     string
	StatusCode int
	Duration   time.Duration
	Errors     []string
	Details    map[string]any
	Values     map[string]any
	Request    *save.Entry
	Response   *save.Entry
}

// ManifestReport groups results for a single manifest.
type ManifestReport struct {
	ManifestID  string
	Namespace   string
	Kind        string
	Name        string
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

// buildReportViewData aggregates results and statistics for the template.
func buildReportViewData(ctx interfaces.ExecutionContext) *ViewData {
	mans := ctx.GetAllManifests()
	results := make([]*save.Result, 0, len(mans))
	mansMap := make(map[string]manifests.Manifest)

	for _, man := range mans {
		key := save.FormSaveKey(man.GetID(), save.ResultKeySuffix)
		if val, ok := ctx.Get(key); ok {
			if res, is := val.([]*save.Result); is {
				results = append(results, res...)
			}
		}
		mansMap[man.GetID()] = man
	}

	if len(results) < 1 {
		return nil
	}

	var reportMap = make(map[string]*ManifestReport)
	var totalCases, passedCases, failedCases int
	var totalTime time.Duration

	for _, res := range results {
		report, ok := reportMap[res.ManifestID]
		if !ok {
			report = &ManifestReport{
				ManifestID: res.ManifestID,
				Target:     res.Target,
				Cases:      []*CaseReport{},
			}
			reportMap[res.ManifestID] = report
		}

		man, ok := mansMap[res.ManifestID]
		if ok {
			report.Namespace = man.GetNamespace()
			report.Name = man.GetName()
			report.Kind = man.GetKind()
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
			Details:    cr.Details,
			Values:     cr.Values,
		}

		report.Cases = append(report.Cases, caseReport)
		report.TotalCases++
		report.TotalTime += cr.Duration

		if cr.Success {
			report.PassedCases++
			passedCases++
		} else {
			report.FailedCases++
			failedCases++
		}

		totalCases++
		totalTime += cr.Duration
	}

	manifestStats := make([]*ManifestReport, 0, len(reportMap))
	for _, m := range reportMap {
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

// ReportGenerator generates an HTML report using html/template and gohtml templates.
type ReportGenerator struct {
	tmpl  *template.Template
	funcs template.FuncMap
}

// NewHTMLReportGenerator creates a new HTMLReportGenerator with custom functions and template directory.
func NewHTMLReportGenerator() (*ReportGenerator, error) {
	funcs := template.FuncMap{
		"formatTime":     func(t time.Time) string { return t.Format("2006-01-02 15:04:05") },
		"formatDuration": func(d time.Duration) string { return d.String() },
		"statusText": func(success bool) string {
			if success {
				return "PASSED"
			}
			return "FAILED"
		},
		"assertText": func(assert string) string {
			if assert == "yes" {
				return "PASSED"
			}
			return "FAILED"
		},
		"float64": func(i int) float64 { return float64(i) },
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"mul": func(a, b float64) float64 { return a * b },
		"prettyJSON": func(v any) string {
			data, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return "<invalid JSON>"
			}
			return string(data)
		},
	}

	tmpl, err := template.New("base.gohtml").Funcs(funcs).ParseFS(htmlTemplates, "templates/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &ReportGenerator{
		tmpl:  tmpl,
		funcs: funcs,
	}, nil
}

// Generate creates an HTML report from test results and writes it to outputPath.
func (g *ReportGenerator) Generate(ctx interfaces.ExecutionContext) error {
	data := buildReportViewData(ctx)

	reportsDir := "reports"
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}
	outputPath := filepath.Join(reportsDir, fmt.Sprintf("report_%s.html", time.Now().Format("2006-01-02-150405")))
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
