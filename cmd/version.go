// Copyright © 2019 Franky Wahl<noreply@example.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

var (
	//GitRevision SHA version at compile time
	GitRevision string

	//Version of the binary
	Version string

	//CreatedAt is when the binary was created
	CreatedAt string
)

// VersionData represents all the information for a version
type VersionData struct {
	CreatedAt time.Time `json:"created_at"`
	Version   string    `json:"version,omitempty"`
	Revision  string    `json:"revision,omitempty"`
}

func newVersionCommand() *cobra.Command {
	// versionCmd represents the version command
	var format string

	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"json": func(i interface{}) (string, error) {
			b, err := json.Marshal(i)
			if err != nil {
				return "", fmt.Errorf("could not marshal data: %w", err)
			}
			return string(b), nil
		},
	}
	cmd := &cobra.Command{
		Use:   "version",
		Short: fmt.Sprintf("Print out %s version information", os.Args[0]),
		Long: fmt.Sprintf(`Get the current version of %s

The -f flag still specifies a format template
applied to a Go struct


type versionData struct {
	CreatedAt time.Time
	Version   string
	Revision  string
}`, os.Args[0]),
		RunE: func(cmd *cobra.Command, args []string) error {
			if CreatedAt == "" {
				CreatedAt = time.Now().UTC().Format(time.RFC3339)
			}
			createdAt, err := time.Parse(time.RFC3339, CreatedAt)
			if err != nil {
				return fmt.Errorf("could not parse time: %w", err)
			}
			data := &VersionData{
				CreatedAt: createdAt.UTC(),
				Version:   Version,
				Revision:  GitRevision,
			}
			tpl, err := template.New("version").Funcs(funcMap).Parse(format + "\n")
			if err != nil {
				return fmt.Errorf("could not create template: %w", err)
			}
			return tpl.Execute(os.Stdout, data)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "{{ . | json }}", "define a format for printing")

	return cmd
}
func init() {
	rootCmd.AddCommand(newVersionCommand())

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
