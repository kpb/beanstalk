package commands

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

//go:embed prime_instructions.md
var defaultPrimeInstructions string

type primeConfig struct {
	Prime *struct {
		Instructions *string `yaml:"instructions"`
	} `yaml:"prime"`
}

func newPrimeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "prime",
		Short: "Print instructions for coding agents",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}

			instructions, err := primeInstructions(workingDirectory)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(command.OutOrStdout(), instructions)
			return err
		},
	}
}

func primeInstructions(workingDirectory string) (string, error) {
	contents, err := os.ReadFile(filepath.Join(workingDirectory, ".beanstalk.yaml"))
	if err != nil {
		if os.IsNotExist(err) {
			return defaultPrimeInstructions, nil
		}
		return "", fmt.Errorf("reading .beanstalk.yaml: %w", err)
	}

	var config primeConfig
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return "", fmt.Errorf("parsing .beanstalk.yaml: %w", err)
	}
	if config.Prime == nil || config.Prime.Instructions == nil {
		return defaultPrimeInstructions, nil
	}
	return *config.Prime.Instructions, nil
}
