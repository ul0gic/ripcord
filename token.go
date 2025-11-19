package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func setTokenInBashrc(token string) error {
	if strings.TrimSpace(token) == "" {
		return errors.New("token cannot be empty")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home dir: %w", err)
	}
	bashrc := filepath.Join(homeDir, ".bashrc")

	content, _ := os.ReadFile(bashrc)
	block := fmt.Sprintf("export DISCORD_TOKEN=\"%s\"", token)
	startMarker := "# >>> ripcord token >>>"
	endMarker := "# <<< ripcord token <<<"

	var builder strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
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
		return err
	}

	if !blockReplaced {
		if len(content) > 0 && content[len(content)-1] != '\n' {
			builder.WriteByte('\n')
		}
		builder.WriteString(startMarker)
		builder.WriteByte('\n')
		builder.WriteString(block)
		builder.WriteByte('\n')
		builder.WriteString(endMarker)
		builder.WriteByte('\n')
	}

	return os.WriteFile(bashrc, []byte(builder.String()), 0o644)
}
