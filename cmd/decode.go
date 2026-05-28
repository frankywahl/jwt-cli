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
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cobra"
)

var errTokenExpired = errors.New("token is expired")

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
	Aliases:      []string{"d"},
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if token == "@-" {
			stdIn, err := readFromStdIn()
			if err != nil {
				return fmt.Errorf("could not read from stdIn: %w", err)
			}
			token = strings.TrimSpace(string(stdIn))
		}

		parser := jwt.NewParser()
		token, err := parser.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return getVerificationKey(token)
		})
		if err != nil {
			if !errors.Is(err, jwt.ErrSignatureInvalid) &&
				!errors.Is(err, jwt.ErrTokenUnverifiable) &&
				!errors.Is(err, jwt.ErrTokenExpired) {
				return err
			}
		}

		result := map[string]interface{}{}
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok {
			cls := map[string]interface{}{}
			for k, v := range claims {
				cls[k] = v
			}
			result["active"] = active(cls["nbf"], cls["exp"])

			for _, tc := range []string{"exp", "nbf", "iat"} { // https://tools.ietf.org/html/rfc7519#section-4.1
				if unix, ok := cls[tc].(float64); ok {
					t := time.Unix(int64(unix), 0)
					cls[tc] = t.UTC().Format(time.RFC3339)
				}
			}
			result["payload"] = cls
		}
		if token.Valid {
			result["signature"] = true
		} else {
			result["signature"] = false
		}
		result["header"] = token.Header

		if err := print(result); err != nil {
			return fmt.Errorf("could not print the result: %w", err)
		}

		if ok && expired(claims) {
			fmt.Fprintln(os.Stderr, "WARNING: token is expired")
			return errTokenExpired
		}
		return nil
	},
}

var token string

func init() {
	decodeCmd.Flags().StringVarP(&secret, "secret", "s", os.Getenv("JWT_SECRET"), "the secret to verify signature (for HMAC) / can use JWT_SECRET env var")
	decodeCmd.Flags().StringVarP(&keyFile, "key-file", "k", "", "path to PEM public key file (for RSA/ECDSA/EdDSA)")
	decodeCmd.Flags().StringVarP(&jwkURL, "jwk-url", "j", "", "URL of a JWK or JWKS endpoint to fetch the verification key from")
	decodeCmd.Flags().StringVarP(&token, "token", "t", "@-", "the token to decode. Using @- will read the token from stdin")
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

func expired(claims jwt.MapClaims) bool {
	exp, ok := claims["exp"].(float64)
	if !ok {
		return false
	}
	return time.Now().After(time.Unix(int64(exp), 0))
}
