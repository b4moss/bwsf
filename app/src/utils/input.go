package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/manifoldco/promptui"
	"golang.org/x/term"
)

// SelectHostType prompts user to select between Cloud and Self-hosted
func SelectHostType() (string, error) {
	prompt := promptui.Select{
		Label: "Bitwarden Cloud or Self-hosted?",
		Items: []string{"cloud", "selfhosted"},
	}

	index, result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to select host type: %w", err)
	}

	_ = index // index is not used but returned by prompt.Run()
	return result, nil
}

// InputURL prompts user to enter self-hosted URL
func InputURL() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	Question("Enter self-hosted URL: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	url := strings.TrimSpace(input)
	if url == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	return url, nil
}

// InputEmail prompts user to enter email address
func InputEmail() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	Question("Enter email address: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	email := strings.TrimSpace(input)
	if email == "" {
		return "", fmt.Errorf("email cannot be empty")
	}

	return email, nil
}

// InputPassword prompts user to enter password (hidden input).
// When BWSF_PASSWORD is set, it is returned without prompting (smoke / automation).
func InputPassword() (string, error) {
	if p := strings.TrimSpace(os.Getenv("BWSF_PASSWORD")); p != "" {
		return p, nil
	}
	return InputSecret("Enter password: ")
}

// InputText prompts for a visible line of text.
func InputText(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	Question("%s", prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	value := strings.TrimSpace(input)
	if value == "" {
		return "", fmt.Errorf("input cannot be empty")
	}
	return value, nil
}

// InputSecret prompts for a secret value (hidden input).
func InputSecret(prompt string) (string, error) {
	Question("%s", prompt)

	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	fmt.Println() // Print newline after hidden input

	password := string(passwordBytes)
	if password == "" {
		return "", fmt.Errorf("input cannot be empty")
	}

	return password, nil
}

// ConfirmOverwrite prompts user to confirm overwrite with y/N
func ConfirmOverwrite(message string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	Question("%s", message)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	response := strings.TrimSpace(strings.ToLower(input))
	return response == "y" || response == "yes", nil
}

// ConfirmYesNo prompts user with a y/N question
// Returns true if user answers "y" or "yes" (case insensitive)
// Returns false for any other input including empty (default is No)
func ConfirmYesNo(message string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	Question("%s", message)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	response := strings.TrimSpace(strings.ToLower(input))
	return response == "y" || response == "yes", nil
}
