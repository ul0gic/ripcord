package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func writeDiscordEnvFile(token string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", errors.New("token cannot be empty")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	envPath := filepath.Join(homeDir, discordEnvFile)

	existing, err := readEnvFile(homeDir)
	if err != nil {
		return "", fmt.Errorf("read existing %s: %w", envPath, err)
	}

	block := fmt.Sprintf("export DISCORD_TOKEN=%q", token)
	const startMarker = "# >>> ripcord token >>>"
	const endMarker = "# <<< ripcord token <<<"

	var builder strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(string(existing)))
	inBlock := false
	blockReplaced := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == startMarker {
			builder.WriteString(startMarker)
			builder.WriteByte('\n')
			builder.WriteString(block)
			builder.WriteByte('\n')
			builder.WriteString(endMarker)
			builder.WriteByte('\n')
			inBlock = true
			blockReplaced = true
			continue
		}
		if trimmed == endMarker {
			inBlock = false
			continue
		}
		if inBlock {
			continue
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	if !blockReplaced {
		if len(existing) > 0 && existing[len(existing)-1] != '\n' {
			builder.WriteByte('\n')
		}
		builder.WriteString(startMarker)
		builder.WriteByte('\n')
		builder.WriteString(block)
		builder.WriteByte('\n')
		builder.WriteString(endMarker)
		builder.WriteByte('\n')
	}

	if err := os.WriteFile(filepath.Clean(envPath), []byte(builder.String()), 0o600); err != nil {
		return "", err
	}
	return envPath, nil
}

func readTokenFromEnvFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	data, err := readEnvFile(homeDir)
	if err != nil || len(data) == 0 {
		return ""
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		eq := strings.IndexByte(line, '=')
		if eq < 0 {
			continue
		}
		if strings.TrimSpace(line[:eq]) != "DISCORD_TOKEN" {
			continue
		}
		val := strings.TrimSpace(line[eq+1:])
		val = strings.Trim(val, `"'`)
		return val
	}
	return ""
}

// readEnvFile loads the ripcord env file via fs.FS so the read path is a
// constant filename relative to the user's home dir. Returns nil bytes (and
// nil error) when the file does not exist.
func readEnvFile(homeDir string) ([]byte, error) {
	data, err := fs.ReadFile(os.DirFS(homeDir), discordEnvFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}
