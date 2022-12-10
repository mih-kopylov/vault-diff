package cmd

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/mih-kopylov/vault-diff/internal/utils"
	"github.com/mih-kopylov/vault-diff/vault"
	"github.com/spf13/cobra"
	"regexp"
	"strconv"
	"strings"
)

func CreateDiffCommand() *cobra.Command {
	flags := struct {
		left  string
		right string
	}{}

	parseSecret := func(value string) (string, int, error) {
		reg, err := regexp.Compile(`(.+):(\d+)`)
		if err != nil {
			return "", 0, err
		}

		submatch := reg.FindStringSubmatch(value)
		if submatch == nil {
			return "", 0, errors.New("malformed secret path: " + value)
		}

		name := submatch[1]
		versionString := submatch[2]
		version, err := strconv.Atoi(versionString)
		if err != nil {
			return "", 0, err
		}

		return name, version, nil
	}

	var result = &cobra.Command{
		Use:   "diff",
		Short: "Shows diff between two provided secret versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := vault.NewClient()
			if err != nil {
				return err
			}

			leftFlagName, leftFlagVersion, err := parseSecret(flags.left)
			if err != nil {
				return err
			}

			rightFlagName, rightFlagVersion, err := parseSecret(flags.right)
			if err != nil {
				return err
			}

			leftSecret, err := vault.GetSecret(client, leftFlagName, leftFlagVersion)
			if err != nil {
				return err
			}

			rightSecret, err := vault.GetSecret(client, rightFlagName, rightFlagVersion)
			if err != nil {
				return err
			}

			edits := myers.ComputeEdits(span.URIFromPath(flags.left), leftSecret, rightSecret)
			diff := fmt.Sprint(gotextdiff.ToUnified(flags.left, flags.right, leftSecret, edits))

			diffLines := strings.Split(diff, "\n")
			addedColor := color.New(color.FgGreen)
			removedColor := color.New(color.FgRed)
			sectionColor := color.New(color.FgCyan)
			defaultColor := color.New()
			for i := 0; i < len(diffLines); i++ {
				var col *color.Color
				diffLine := diffLines[i]
				if strings.HasPrefix(diffLine, "+") {
					col = addedColor
				} else if strings.HasPrefix(diffLine, "-") {
					col = removedColor
				} else if strings.HasPrefix(diffLine, "@@") {
					col = sectionColor
				} else {
					col = defaultColor
				}
				_, _ = col.Println(diffLine)
			}

			return nil
		},
	}

	result.Flags().StringVarP(&flags.left, "left", "l", "", "Left side of the key to compare. Use 'key:version' format")
	utils.MarkFlagRequiredOrFail(result.Flags(), "left")

	result.Flags().StringVarP(&flags.right, "right", "r", "", "Right side of the key to compare. Use 'key:version' format")
	utils.MarkFlagRequiredOrFail(result.Flags(), "right")

	return result
}
