package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pkg/errors"
)

var update = flag.Bool("update", false, "update golden files")

const binaryName = "jwt-cli"

var binaryPath string

func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		fmt.Printf("could not change dir: %v", err)
		os.Exit(1)
	}
	make := exec.Command("go", "build", "-o", binaryName)
	if err := make.Run(); err != nil {
		fmt.Printf("could not make binary for %s: %v", binaryName, err)
		os.Exit(1)
	}

	abs, err := filepath.Abs(binaryName)
	if err != nil {
		fmt.Printf("could not get abs path for %s: %v", binaryName, err)
		os.Exit(1)
	}

	if err := os.Chdir(filepath.Dir(abs)); err != nil {
		fmt.Printf("could not change dir: %v", err)
		os.Exit(1)
	}

	defer os.Exit(m.Run())
	defer os.Remove(binaryName)
}

func TestCLI(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"no-arguments", []string{}, "no-args"},

		// Encoding
		{"encoding", []string{"encode", "--help"}, "encoding-help"},
		{"encoding with data", []string{"encode", "-d", "{\"hello\":\"world\",\"exp\":0,\"nbf\":0,\"iat\":0}"}, "encoding"},
		{"encoding with secret", []string{"encode", "--secret", "SECRET", "-d", "{\"hello\":\"world\",\"exp\":0,\"nbf\":0,\"iat\":0}"}, "encoding-secret"},
		{"encoding with sign method", []string{"encode", "--secret", "SECRET", "--sign-method", "H512", "-d", "{\"hello\":\"world\",\"exp\":0,\"nbf\":0,\"iat\":0}"}, "encoding-secret-sign-method"},

		// Decoding
		{"decode", []string{"decode", "--help"}, "decode-help"},
		{"decode token", []string{"decode", "-t", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImhlbGxvIjoid29ybGQiLCJpYXQiOjAsIm5iZiI6MH0.nah_yz6eXP9Vu0W-ksnCc30eqpXiHwepssBYePjdJUo"}, "decode-no-args"},
		{"decode token with secret", []string{"decode", "-t", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImhlbGxvIjoid29ybGQiLCJpYXQiOjAsIm5iZiI6MH0.nah_yz6eXP9Vu0W-ksnCc30eqpXiHwepssBYePjdJUo", "--secret", "SECRET"}, "decode-with-secret"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			dir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("./"+binaryName, tt.args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatal(err)
			}

			fixturePath := fmt.Sprintf("%s/integration/fixtures/%s", dir, tt.fixture)

			if *update {
				writeFixture(t, fixturePath, string(output))
			}
			actual := string(output)
			expected, err := loadFixture(t, fixturePath)
			if err != nil {
				t.Fatalf("could not load fixture: %v", err)
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("\nactual:\n%s\nexpected:\n%s", actual, expected)
			}
		})
	}
}

func writeFixture(t *testing.T, name string, content string) error {
	return ioutil.WriteFile(name, []byte(content), 0644)
}
func loadFixture(t *testing.T, path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.Wrapf(err, "could not read file %s", path)
	}
	return string(content), nil
}
