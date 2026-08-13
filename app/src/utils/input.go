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

	Question("Enter password: ")

	// Read password without echoing to terminal
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	fmt.Println() // Print newline after password input

	password := string(passwordBytes)
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
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

// CleanMismatchAction labels returned by SelectCleanMismatchAction.
const (
	CleanMismatchAbort                    = "abort"
	CleanMismatchOverwriteRemoteThenClean = "overwrite_remote_then_clean"
	CleanMismatchRemoveLocal              = "remove_local"
)

// SelectCleanMismatchAction prompts for a single action when local/remote contents differ.
// Options are presented as a radio-style single select (Abort first for safety).
func SelectCleanMismatchAction(mismatchedFiles []string) (string, error) {
	Warningln("[bwsf Alert] File contents on remote are mismatch.")
	if len(mismatchedFiles) > 0 {
		Infoln("[INFO] Mismatched file(s):")
		for _, name := range mismatchedFiles {
			Infoln("  -", name)
		}
	}
	Infoln("Are you sure to remove files from local?")

	items := []string{
		"Abort",
		"Overwrite remote with my local, then clean",
		"Remove local without updating remote (DANGER)",
	}

	prompt := promptui.Select{
		Label: "Select an action",
		Items: items,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return CleanMismatchAbort, fmt.Errorf("failed to select clean action: %w", err)
	}

	switch index {
	case 0:
		return CleanMismatchAbort, nil
	case 1:
		return CleanMismatchOverwriteRemoteThenClean, nil
	case 2:
		return CleanMismatchRemoveLocal, nil
	default:
		return CleanMismatchAbort, nil
	}
}
