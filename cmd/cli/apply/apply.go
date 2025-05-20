package apply

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/apiqube/cli/internal/core/io"
	"github.com/apiqube/cli/internal/core/manifests"
	"github.com/apiqube/cli/internal/core/store"
	"github.com/apiqube/cli/internal/validate"
	"github.com/apiqube/cli/ui/cli"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	Cmd.Flags().StringP("file", "f", ".", "Path to manifests file, by default is current")
}

var Cmd = &cobra.Command{
	Use:           "apply",
	Short:         "Apply resources from manifest files",
	Long:          "Apply configuration from YAML manifests with validation and version control",
	SilenceErrors: true,
	SilenceUsage:  true,
	Run: func(cmd *cobra.Command, args []string) {
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			cli.Errorf("Failed to parse input file flag: %v", err)
			return
		}

		cli.Infof("Loading manifests from: %s", file)
		loadedMans, cachedMans, err := io.LoadManifests(file)
		if err != nil {
			cli.Errorf("Critical load error:\n%s", formatLoadError(err, file))
			return
		}

		cli.Info("Validating manifests...")
		validator := validate.NewManifestValidator(validate.NewValidator(), cli.Instance())

		validator.Validate(loadedMans...)

		validMans := validator.Valid()
		if len(validMans) == 0 {
			cli.Warning("No valid manifests to apply")
			return
		}

		printManifestsLoadResult(validMans, cachedMans)

		cli.Infof("Saving %d manifests to storage...", len(validMans))
		if err := store.Save(validMans...); err != nil {
			cli.Errorf("Storage error: -\n%s", err.Error())
			return
		}

		printPostApplySummary(validMans)
		cli.Successf("Successfully applied %d manifests", len(validMans))
	},
}

func formatLoadError(err error, file string) string {
	if os.IsNotExist(err) {
		return fmt.Sprintf("File not found: \n%s- Please check the path and try again", file)
	}
	var yamlErr *yaml.TypeError
	if errors.As(err, &yamlErr) {
		return fmt.Sprintf("YAML syntax error:\n%s", indentYAMLError(yamlErr))
	}

	return err.Error()
}

func printManifestsLoadResult(newMans, cachedMans []manifests.Manifest) {
	if len(newMans) > 0 {
		var builder strings.Builder

		for _, m := range newMans {
			builder.WriteString(fmt.Sprintf("\n- %s %s",
				m.GetID(),
				fmt.Sprintf("(h: %s)", cli.ShortHash(m.GetMeta().GetHash())),
			))
		}

		cli.Infof("New manifests detected: %s", builder.String())
	}

	if len(cachedMans) > 0 {
		var builder strings.Builder

		for _, m := range cachedMans {
			builder.WriteString(fmt.Sprintf("\n- %s %s",
				m.GetID(),
				fmt.Sprintf("(h: %s)", cli.ShortHash(m.GetMeta().GetHash())),
			))

			cli.Infof("Using cached manifest: %s", builder.String())
		}
	}
}

func printPostApplySummary(mans []manifests.Manifest) {
	stats := make(map[string]int)
	for _, m := range mans {
		stats[m.GetKind()]++
	}

	var builder strings.Builder

	for kind, count := range stats {
		builder.WriteString(fmt.Sprintf("\n- %s: %d", kind, count))
	}

	cli.Infof("Applied manifests by kind: %s", builder.String())
}

func indentYAMLError(err *yaml.TypeError) string {
	return "  " + strings.Join(err.Errors, "\n  ")
}
