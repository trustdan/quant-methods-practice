package tutor

import (
	"fmt"
	"strings"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/storage"
)

// ExportNoteToMarkdown formats a saved explanation as clean, portable UTF-8 Markdown
// compatible with Typora and Obsidian, including YAML frontmatter and advisory notices.
func ExportNoteToMarkdown(note *storage.SavedExplanation, scenarioContext string) string {
	if note == nil {
		return ""
	}

	var sb strings.Builder

	title := note.ProviderInfo.Title
	if strings.TrimSpace(title) == "" {
		title = fmt.Sprintf("%s - Explanation", note.Topic)
	}

	// 1. YAML frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("id: %q\n", note.ID))
	sb.WriteString(fmt.Sprintf("title: %q\n", title))
	sb.WriteString(fmt.Sprintf("topic: %q\n", note.Topic))
	if len(note.ProviderInfo.Concepts) > 0 {
		sb.WriteString("concepts:\n")
		for _, c := range note.ProviderInfo.Concepts {
			sb.WriteString(fmt.Sprintf("  - %q\n", c))
		}
	}
	provider := note.ProviderInfo.Provider
	if provider == "" {
		provider = "offline"
	}
	sb.WriteString(fmt.Sprintf("provider: %q\n", provider))
	if note.ProviderInfo.Model != "" {
		sb.WriteString(fmt.Sprintf("model: %q\n", note.ProviderInfo.Model))
	}
	sb.WriteString("advisory: \"Advisory / Self-study reference\"\n")
	sb.WriteString(fmt.Sprintf("created_at: %q\n", note.CreatedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("updated_at: %q\n", note.UpdatedAt.Format(time.RFC3339)))
	sb.WriteString("---\n\n")

	// 2. Main title and advisory alert
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))
	sb.WriteString("> [!NOTE]\n")
	sb.WriteString("> **Advisory Note**: This explanation is for self-study and reference. It does not replace official course guidelines or modify graded evidence.\n\n")

	// 3. Metadata context block
	sb.WriteString("## Context\n\n")
	sb.WriteString(fmt.Sprintf("- **Topic**: %s\n", note.Topic))
	if len(note.ProviderInfo.Concepts) > 0 {
		sb.WriteString(fmt.Sprintf("- **Concepts**: %s\n", strings.Join(note.ProviderInfo.Concepts, ", ")))
	}
	sb.WriteString(fmt.Sprintf("- **Provider**: %s\n", provider))
	sb.WriteString(fmt.Sprintf("- **Saved At**: %s\n", note.UpdatedAt.Format(time.RFC3339)))
	sb.WriteString("\n")

	// 4. Problem scenario context if available
	if strings.TrimSpace(scenarioContext) != "" {
		sb.WriteString("### Problem Context\n\n")
		sb.WriteString(strings.TrimSpace(scenarioContext))
		sb.WriteString("\n\n")
	}

	// 5. Main explanation content with LaTeX math
	sb.WriteString("## Explanation\n\n")
	sb.WriteString(strings.TrimSpace(note.RawMarkdown))
	sb.WriteString("\n\n")

	// 6. Footer
	sb.WriteString("---\n")
	sb.WriteString("*Exported from Quant Methods Practice. Mathematical notation formatted in standard LaTeX ($...$ and $$...$$) compatible with MathJax, Obsidian, and Typora.*\n")

	return sb.String()
}
