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
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/spf13/cobra"
)

// encodeCmd represents the encode command
var encodeCmd = &cobra.Command{
	Use:   "encode",
	Short: "Encode data using a secret",
	Long: fmt.Sprintf(`Encode data for getting a JWT token

This takes in options for stuff such as secret
Data is read from the output

Example:
  echo '{"Hello": "World"}' | %s encode --secret SECRET
`, os.Args[0]),
	Aliases: []string{"e"},
	RunE: func(cmd *cobra.Command, args []string) error {
		var signAlgorithm *jwt.SigningMethodHMAC
		switch signMethod {
		case "H256":
			signAlgorithm = jwt.SigningMethodHS256
		case "H384":
			signAlgorithm = jwt.SigningMethodHS384
		case "H512":
			signAlgorithm = jwt.SigningMethodHS512
		}

		if data == "@-" {
			stdIn, err := readFromStdIn()
			if err != nil {
				return fmt.Errorf("could not read from stdIn: %w", err)
			}
			data = string(stdIn)
		}

		var dataJSON map[string]interface{}
		if err := json.Unmarshal([]byte(data), &dataJSON); err != nil {
			return fmt.Errorf("could not unmarshal the data: %w", err)
		}

		if t, ok := dataJSON["exp"]; ok { // t is a unix timestamp
			dataJSON["exp"] = t
		} else {
			dataJSON["exp"] = time.Now().Add(5 * 24 * time.Hour).Unix()
		}

		if t, ok := dataJSON["iat"]; ok { // t is a unix timestamp
			dataJSON["iat"] = t
		} else {
			dataJSON["iat"] = time.Now().Unix()
		}

		claim := jwt.NewWithClaims(
			signAlgorithm, jwt.MapClaims(
				dataJSON,
			),
		)

		token, err := claim.SignedString([]byte(secret))
		if err != nil {
			return fmt.Errorf("could not write token")
		}

		fmt.Printf("%s\n", token)
		return nil
	},
}

var data string

func init() {
	encodeCmd.Flags().StringVarP(&secret, "secret", "s", os.Getenv("JWT_SECRET"), "the secret needed")
	encodeCmd.Flags().StringVarP(&signMethod, "sign-method", "m", "H256", "the signing method")
	encodeCmd.Flags().StringVarP(&data, "data", "d", "@-", "the claims to be signed. Using @- will read the data from stdin")
	rootCmd.AddCommand(encodeCmd)
}
