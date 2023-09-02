/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andreyvit/diff"
	"github.com/kbudde/k8n/internal/ytt"
	"github.com/spf13/cobra"
)

// testCmd represents the test command.
//
//nolint:exhaustruct,gochecknoglobals
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run test cases",
	Long:  `Run test cases using fixtures and the configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		input, err := cmd.Flags().GetString("input")
		cobra.CheckErr(err)
		folder, err := cmd.Flags().GetString("ytt")
		cobra.CheckErr(err)
		if folder == "" {
			folder = filepath.Dir(input)
		}
		testfolder, err := cmd.Flags().GetString("tests")
		cobra.CheckErr(err)
		if testfolder == "" {
			testfolder = filepath.Dir(input)
		}

		tests, err := getTests(testfolder)
		cobra.CheckErr(err)
		if len(tests) == 0 {
			fmt.Println("No tests found in ", testfolder)
			os.Exit(1)
		}
		for _, test := range tests {
			input := test.input
			name := getTestName(input)
			out, err := ytt.Render(input, folder)
			if err != nil {
				fmt.Printf("%s failed:\n%v", name, err)
				os.Exit(1)
			}

			diff := compare(test.Output(), out)
			if diff != nil {
				fmt.Printf("❌ %s failed.\n", name)
				outputDiff(diff)

				continue
			}
			fmt.Println("✅", name, "passed")
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().String("input", "input.yaml", "path to data file. See `k8n read`")
	testCmd.Flags().String("ytt", "", "path to ytt files. Defaults to the directory of the input file.")
	testCmd.Flags().String("tests", "",
		"folder with test files. Pattern input.XYZ.yaml and output.XYZ.yaml. Defaults to the directory of the input file.")
}

type test struct {
	input  string
	output string
}

func getTests(folder string) ([]test, error) {
	// find all files matching input.*.yaml
	inputs, err := filepath.Glob(filepath.Join(folder, "input.*.yaml"))
	if err != nil {
		return nil, err
	}

	tests := make([]test, 0, len(inputs))

	for _, input := range inputs {
		name := getTestName(input)
		output := filepath.Join(folder, fmt.Sprintf("output.%s.yaml", name))
		// check if output file exists
		if _, err := os.Stat(output); os.IsNotExist(err) {
			return nil, fmt.Errorf("output file '%s' does not exist", output)
		}

		tests = append(tests, test{input: input, output: output})
	}

	return tests, nil
}

// Output returns the content of the output file.
func (t test) Output() string {
	b, err := os.ReadFile(t.output)
	if err != nil {
		panic(err)
	}

	return string(b)
}

// expected and actual are yaml files containing multiple documents.
func compare(expected string, actual []byte) []string {
	actualStr := string(actual)
	if a, e := strings.TrimSpace(expected), strings.TrimSpace(actualStr); a != e {
		return diff.LineDiffAsLines(a, e)
	}

	return nil
}

func getTestName(input string) string {
	name := filepath.Base(input)
	name = strings.TrimPrefix(name, "input.")
	name = strings.TrimSuffix(name, ".yaml")
	name = strings.TrimSpace(name)

	return name
}

func outputDiff(diff []string) {
	for _, line := range diff {
		if line[0] == '+' {
			fmt.Printf("\x1b[32m%s\x1b[0m\n", line)

			continue
		}

		if line[0] == '-' {
			fmt.Printf("\x1b[31m%s\x1b[0m\n", line)

			continue
		}

		fmt.Println(line)
	}
}
