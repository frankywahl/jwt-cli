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
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"os"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var secret, signMethod, keyFile, jwkURL string

// resolveSigningMethod maps a short method name to a jwt.SigningMethod.
func resolveSigningMethod(method string) (jwt.SigningMethod, error) {
	switch method {
	case "H256":
		return jwt.SigningMethodHS256, nil
	case "H384":
		return jwt.SigningMethodHS384, nil
	case "H512":
		return jwt.SigningMethodHS512, nil
	case "R256":
		return jwt.SigningMethodRS256, nil
	case "R384":
		return jwt.SigningMethodRS384, nil
	case "R512":
		return jwt.SigningMethodRS512, nil
	case "E256":
		return jwt.SigningMethodES256, nil
	case "E384":
		return jwt.SigningMethodES384, nil
	case "E512":
		return jwt.SigningMethodES512, nil
	case "EdDSA":
		return jwt.SigningMethodEdDSA, nil
	default:
		return nil, fmt.Errorf("unsupported signing method: %s", method)
	}
}

// readKeyFile reads a PEM key file from disk.
func readKeyFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read key file %s: %w", path, err)
	}
	return data, nil
}

// getSigningKey returns the appropriate signing key for the given method.
func getSigningKey(method jwt.SigningMethod) (interface{}, error) {
	switch method.(type) {
	case *jwt.SigningMethodHMAC:
		return []byte(secret), nil
	case *jwt.SigningMethodRSA:
		if keyFile == "" {
			return nil, fmt.Errorf("RSA signing requires --key-file with a PEM private key")
		}
		pem, err := readKeyFile(keyFile)
		if err != nil {
			return nil, err
		}
		return jwt.ParseRSAPrivateKeyFromPEM(pem)
	case *jwt.SigningMethodECDSA:
		if keyFile == "" {
			return nil, fmt.Errorf("ECDSA signing requires --key-file with a PEM private key")
		}
		pem, err := readKeyFile(keyFile)
		if err != nil {
			return nil, err
		}
		return jwt.ParseECPrivateKeyFromPEM(pem)
	case *jwt.SigningMethodEd25519:
		if keyFile == "" {
			return nil, fmt.Errorf("EdDSA signing requires --key-file with a PEM private key")
		}
		pem, err := readKeyFile(keyFile)
		if err != nil {
			return nil, err
		}
		return jwt.ParseEdPrivateKeyFromPEM(pem)
	default:
		return nil, fmt.Errorf("unsupported signing method type: %v", method.Alg())
	}
}

// getKeyFromJWKURL fetches a JWK or JWKS from the given URL and returns the
// raw key material matching the token's "kid" header (if present). If the
// endpoint returns a single JWK object it is wrapped into a set automatically.
func getKeyFromJWKURL(token *jwt.Token, url string) (interface{}, error) {
	set, err := jwk.Fetch(context.Background(), url)
	if err != nil {
		return nil, fmt.Errorf("could not fetch JWK(S) from %s: %w", url, err)
	}

	// Try to match by "kid" if the token carries one.
	var key jwk.Key
	if kid, ok := token.Header["kid"].(string); ok && kid != "" {
		key, ok = set.LookupKeyID(kid)
		if !ok {
			return nil, fmt.Errorf("no key with kid %q found at %s", kid, url)
		}
	} else {
		if set.Len() == 0 {
			return nil, fmt.Errorf("JWK set at %s is empty", url)
		}
		var ok bool
		key, ok = set.Key(0)
		if !ok {
			return nil, fmt.Errorf("could not retrieve first key from JWK set at %s", url)
		}
	}

	var raw interface{}
	if err := key.Raw(&raw); err != nil {
		return nil, fmt.Errorf("could not extract raw key from JWK: %w", err)
	}

	// golang-jwt expects the public-key type for verification.
	switch k := raw.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey, nil
	case *ecdsa.PrivateKey:
		return &k.PublicKey, nil
	case ed25519.PrivateKey:
		return k.Public(), nil
	default:
		return raw, nil
	}
}

// getVerificationKey returns the appropriate verification key for the token's method.
func getVerificationKey(token *jwt.Token) (interface{}, error) {
	if jwkURL != "" {
		return getKeyFromJWKURL(token, jwkURL)
	}
	switch token.Method.(type) {
	case *jwt.SigningMethodHMAC:
		return []byte(secret), nil
	case *jwt.SigningMethodRSA:
		if keyFile == "" {
			return nil, fmt.Errorf("RSA verification requires --key-file with a PEM public key")
		}
		pem, err := readKeyFile(keyFile)
		if err != nil {
			return nil, err
		}
		key, err := jwt.ParseRSAPublicKeyFromPEM(pem)
		if err != nil {
			// Try parsing as private key and extract public key
			privKey, privErr := jwt.ParseRSAPrivateKeyFromPEM(pem)
			if privErr != nil {
				return nil, err
			}
			return &privKey.PublicKey, nil
		}
		return key, nil
	case *jwt.SigningMethodECDSA:
		if keyFile == "" {
			return nil, fmt.Errorf("ECDSA verification requires --key-file with a PEM public key")
		}
		pem, err := readKeyFile(keyFile)
		if err != nil {
			return nil, err
		}
		key, err := jwt.ParseECPublicKeyFromPEM(pem)
		if err != nil {
			privKey, privErr := jwt.ParseECPrivateKeyFromPEM(pem)
			if privErr != nil {
				return nil, err
			}
			return &privKey.PublicKey, nil
		}
		return key, nil
	case *jwt.SigningMethodEd25519:
		if keyFile == "" {
			return nil, fmt.Errorf("EdDSA verification requires --key-file with a PEM public key")
		}
		pem, err := readKeyFile(keyFile)
		if err != nil {
			return nil, err
		}
		key, err := jwt.ParseEdPublicKeyFromPEM(pem)
		if err != nil {
			privKey, privErr := jwt.ParseEdPrivateKeyFromPEM(pem)
			if privErr != nil {
				return nil, err
			}
			// ed25519.PrivateKey contains the public key in the last 32 bytes
			if edKey, ok := privKey.(ed25519.PrivateKey); ok {
				return edKey.Public(), nil
			}
			return nil, err
		}
		return key, nil
	default:
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
}

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
