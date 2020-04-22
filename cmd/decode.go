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
	"fmt"
	"os"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/spf13/cobra"
)

// decodeCmd represents the decode command
var decodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "Decode a JWT token",
	Long: `Decode JWT web tokens.

This will print out the result of decoding the token. This will consist
of a 3 part response:

	* active: is the token enabled for the current time
	* header: information encoded from the standard
	* payload: the set of claims the token contains
	* signature: whether the token has been signed or not
`,
	Aliases: []string{"d"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if token == "@-" {
			stdIn, err := readFromStdIn()
			if err != nil {
				return fmt.Errorf("could not read from stdIn: %w", err)
			}
			token = strings.TrimSpace(string(stdIn))
		}

		token, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(secret), nil
		})

		if err != nil {
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
					return fmt.Errorf("token malformed")
				}
			}
		}

		result := make([]map[string]interface{}, 2)

		// result := map[string]interface{}{}
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok {
			cls := map[string]interface{}{}
			for k, v := range claims {
				cls[k] = v
			}

			for _, tc := range []string{"exp", "nbf", "iat"} { // https://tools.ietf.org/html/rfc7519#section-4.1
				if unix, ok := cls[tc].(float64); ok {
					t := time.Unix(int64(unix), 0)
					now := time.Now()
					if now.Before(t) && tc == "nbf" && !skipTime {
						return fmt.Errorf("token is not valid for another %s, use `--skip-time`", t.Sub(now))
					}
					if now.After(t) && tc == "exp" && !skipTime {
						return fmt.Errorf("token has expired for %s, use `--skip-time`", now.Sub(t))
					}
					if tc == "iat" {
						cls[tc] = t.UTC().Format(time.RFC3339)
					}
				}
			}
			result[1] = cls
		}
		if !token.Valid && !skipSignature { // This also looks at nbf and exp
			return fmt.Errorf("wrong signature, use `--skip-signature` if you dont want sign")
		}
		result[0] = token.Header

		if err := print(result); err != nil {
			return fmt.Errorf("could not print the result: %w", err)
		}
		return nil
	},
}

var token string
var skipSignature bool
var skipTime bool

func init() {
	decodeCmd.Flags().StringVarP(&secret, "secret", "s", os.Getenv("JWT_SECRET"), "the secret to verify signature / can use JWT_SECRET env var")
	decodeCmd.Flags().StringVarP(&token, "token", "t", "@-", "the token to decode. Using @- will read the token from stdin")
	decodeCmd.Flags().BoolVarP(&skipSignature, "skip-signature", "", false, "skip signature validation")
	decodeCmd.Flags().BoolVarP(&skipTime, "skip-time", "", false, "skip time based validation (nbf, exp)")
	rootCmd.AddCommand(decodeCmd)
}

func active(nbf, exp interface{}) bool {
	if unix, ok := nbf.(float64); ok {
		t := time.Unix(int64(unix), 0)
		if time.Now().Before(t) {
			return false
		}
	}
	if unix, ok := exp.(float64); ok {
		t := time.Unix(int64(unix), 0)
		if time.Now().After(t) {
			return false
		}
	}
	return true
}
