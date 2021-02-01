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
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var secret, signMethod string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("%s", os.Args[0]),
	Short: "JWT cli for generating web tokens",
	Long: `JWT cli allows you to interact with JWT (JSON Web Tokens) directly
via the command line.

https://tools.ietf.org/html/rfc7519
`,
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jwt-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".jwt-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".jwt-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// readFromStdIn is here to read data from a pipe
// https://flaviocopes.com/go-shell-pipes/
func readFromStdIn() ([]byte, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return []byte{}, fmt.Errorf("could not get StdIn")
	}
	if info.Mode()&os.ModeCharDevice == os.ModeCharDevice { // || info.Size() <= 0 {
		return []byte{}, fmt.Errorf("could not read the data")
	}

	reader := bufio.NewReader(os.Stdin)
	var output []byte

	for {
		input, err := reader.ReadByte()
		if err != nil && err == io.EOF {
			break
		}
		output = append(output, input)
	}

	return output, nil
}

// print outputs a JSON marshal dump to the output
func print(result map[string]interface{}) error {
	output, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("could not marhsal result: %w", err)
	}
	fmt.Println(string(output))
	return nil
}
