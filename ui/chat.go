package ui

import (
	"strings"
	"time"
)

// ChatMessage is rendered in the center chat panel.
type ChatMessage struct {
	Role      string
	Content   string
	Timestamp time.Time
}

func formatChat(messages []ChatMessage, width int) string {
	if len(messages) == 0 {
		return "No conversation yet.\n\nType a prompt below and press Enter to talk to AWaN Core."
	}

	var builder strings.Builder
	for i, message := range messages {
		if i > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(strings.ToUpper(message.Role))
		builder.WriteString(" · ")
		builder.WriteString(message.Timestamp.Format("15:04:05"))
		builder.WriteString("\n")
		builder.WriteString(wrapText(message.Content, width))
	}

	return builder.String()
}

func formatMemory(snapshot memoryView, width int) string {
	var builder strings.Builder
	builder.WriteString("SHORT TERM\n")
	if len(snapshot.ShortTerm) == 0 {
		builder.WriteString("No short-term memory.\n")
	} else {
		for _, entry := range snapshot.ShortTerm {
			builder.WriteString("- ")
			builder.WriteString(entry.Role)
			builder.WriteString(": ")
			builder.WriteString(wrapText(entry.Content, width-4))
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\nLONG TERM\n")
	if len(snapshot.LongTerm) == 0 {
		builder.WriteString("No long-term memory.\n")
	} else {
		for _, entry := range snapshot.LongTerm {
			builder.WriteString("- ")
			builder.WriteString(entry.Role)
			builder.WriteString(": ")
			builder.WriteString(wrapText(entry.Content, width-4))
			builder.WriteString("\n")
		}
	}

	return strings.TrimSpace(builder.String())
}

func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var (
		line    strings.Builder
		builder strings.Builder
	)

	for _, word := range words {
		if line.Len() == 0 {
			line.WriteString(word)
			continue
		}

		if line.Len()+1+len(word) > width {
			builder.WriteString(line.String())
			builder.WriteString("\n")
			line.Reset()
			line.WriteString(word)
			continue
		}

		line.WriteString(" ")
		line.WriteString(word)
	}

	if line.Len() > 0 {
		builder.WriteString(line.String())
	}

	return builder.String()
}
