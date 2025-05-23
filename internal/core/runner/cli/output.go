package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/apiqube/cli/ui"

	"github.com/apiqube/cli/ui/cli"

	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
)

var _ interfaces.Output = (*Output)(nil)

type Output struct{}

func NewOutput() *Output {
	return &Output{}
}

func (o *Output) StartCase(manifest manifests.Manifest, caseName string) {
	cli.LogStyledf(ui.TypeInfo,
		"Start [%s] case from [%s] manifest",
		cli.LogPair{Message: caseName, Style: &cli.InfoStyle},
		cli.LogPair{Message: manifest.GetName(), Style: &cli.WarningStyle},
	)
}

func (o *Output) EndCase(manifest manifests.Manifest, caseName string, result *interfaces.CaseResult) {
	if result != nil {
		successStyle := cli.SuccessStyle
		successText := "yes"
		if !result.Success {
			successStyle = cli.ErrorStyle
			successText = "no"
		}

		errorsFormatted := ""
		if len(result.Errors) > 0 {
			var errorsBuilder strings.Builder
			for _, err := range result.Errors {
				errorsBuilder.WriteString(fmt.Sprintf("\n- %s", err))
			}
			errorsFormatted = fmt.Sprintf("\nErrors: %s", errorsBuilder.String())
		}

		detailsFormatted := ""
		if len(result.Details) > 0 {
			var detailsBuilder strings.Builder
			var keys []string

			for key := range result.Details {
				keys = append(keys, key)
			}

			sort.Strings(keys)

			for _, key := range keys {
				detailsBuilder.WriteString(fmt.Sprintf("\n- %s: %v", key, result.Details[key]))
			}

			detailsFormatted = fmt.Sprintf("\nDetails: %s", detailsBuilder.String())
		}

		cli.LogStyledf(
			ui.TypeInfo,
			"Finish [%s] case from [%s] manifest with next results\n"+
				"Result: %s\n"+
				"Success: %s\n"+
				"Status Code: %s\n"+
				"Duration: %s%s%s",
			cli.LogPair{Message: caseName, Style: &cli.InfoStyle},
			cli.LogPair{Message: manifest.GetName(), Style: &cli.WarningStyle},
			cli.LogPair{Message: result.Name},
			cli.LogPair{Message: successText, Style: &successStyle},
			cli.LogPair{Message: fmt.Sprint(result.StatusCode)},
			cli.LogPair{Message: result.Duration.String()},
			cli.LogPair{Message: errorsFormatted, Style: &cli.ErrorStyle},
			cli.LogPair{Message: detailsFormatted, Style: &cli.TimestampStyle},
		)
	} else {
		cli.LogStyledf(ui.TypeInfo,
			"Finish [%s] case from [%s] manifest with next results",
			cli.LogPair{Message: caseName, Style: &cli.InfoStyle},
			cli.LogPair{Message: manifest.GetName(), Style: &cli.WarningStyle},
		)
	}
}

func (o *Output) ReceiveMsg(msg any) {
	cli.Infof("Receiving message %s", msg)
}

func (o *Output) Log(level interfaces.LogLevel, msg string) {
	switch level {
	case interfaces.DebugLevel:
		cli.Debug(msg)
	case interfaces.InfoLevel:
		cli.Info(msg)
	case interfaces.WarnLevel:
		cli.Warning(msg)
	case interfaces.ErrorLevel:
		cli.Error(msg)
	case interfaces.FatalLevel:
		cli.Fatal(msg)
	default:
		cli.Info(msg)
	}
}

func (o *Output) Logf(level interfaces.LogLevel, format string, args ...any) {
	switch level {
	case interfaces.DebugLevel:
		cli.Debugf(format, args...)
	case interfaces.InfoLevel:
		cli.Infof(format, args...)
	case interfaces.WarnLevel:
		cli.Warningf(format, args...)
	case interfaces.ErrorLevel:
		cli.Errorf(format, args...)
	case interfaces.FatalLevel:
		cli.Fatalf(format, args...)
	default:
		cli.Infof(format, args...)
	}
}

func (o *Output) DumpValues(values map[string]any) {
	if values != nil {
		var rows []string
		for k, v := range values {
			rows = append(rows, fmt.Sprintf("%v: %v", k, v))
		}

		cli.Printf("Damping values: \n%s", strings.Join(rows, "\n"))
	}
}

func (o *Output) Error(err error) {
	cli.Error(err.Error())
}
