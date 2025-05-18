package cli

import (
	"fmt"
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/runner/interfaces"
	"github.com/apiqube/cli/ui"
	"strings"
)

var _ interfaces.Output = (*Output)(nil)

type Output struct{}

func NewOutput() *Output {
	return &Output{}
}

func (o *Output) StartCase(manifest manifests.Manifest, caseName string) {
	ui.Infof("Start %s case from %s manifest", caseName, manifest.GetName())
}

func (o *Output) EndCase(manifest manifests.Manifest, caseName string, result *interfaces.CaseResult) {
	if result != nil {
		ui.Println(fmt.Sprintf(
			`Finish %s case from %s manifest with next reults
					Result:  	   %s
 					Success:	   %v
					Status Code:   %d
					Duration: 	   %s`,
			caseName,
			manifest.GetName(),
			result.Name,
			result.Success,
			result.StatusCode,
			result.Duration.String(),
		),
		)
	} else {
		ui.Infof("Finish %s case from %s manifest", caseName, manifest.GetName())
	}
}

func (o *Output) ReceiveMsg(msg any) {
	ui.Infof("Receiving message %s", msg)
}

func (o *Output) Log(level interfaces.LogLevel, msg string) {
	switch level {
	case interfaces.DebugLevel:
		ui.Debug(msg)
	case interfaces.InfoLevel:
		ui.Info(msg)
	case interfaces.WarnLevel:
		ui.Warning(msg)
	case interfaces.ErrorLevel:
		ui.Error(msg)
	default:
		ui.Info(msg)
	}
}

func (o *Output) Logf(level interfaces.LogLevel, format string, args ...any) {
	switch level {
	case interfaces.DebugLevel:
		ui.Debugf(format, args...)
	case interfaces.InfoLevel:
		ui.Infof(format, args...)
	case interfaces.WarnLevel:
		ui.Warningf(format, args...)
	case interfaces.ErrorLevel:
		ui.Errorf(format, args...)
	default:
		ui.Infof(format, args...)
	}
}

func (o *Output) DumpValues(values map[string]any) {
	if values != nil {
		var rows []string
		for k, v := range values {
			rows = append(rows, fmt.Sprintf("%v: %v", k, v))
		}

		ui.Printf("Damping values: \n%s", strings.Join(rows, "\n"))
	}
}

func (o *Output) Error(err error) {
	ui.Error(err.Error())
}
