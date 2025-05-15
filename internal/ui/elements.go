package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func Progress(percent float64, text ...string) {
	if !IsEnabled() {
		return
	}

	instance.queueUpdate(func(m *uiModel) {
		textStr := ""
		if len(text) > 0 {
			textStr = strings.Join(text, " ")
		}

		if percent >= 100 {
			m.removeLastProgress()
			return
		}

		m.updateOrAddProgress(percent, textStr)
	})
}

func Loader(show bool, text ...string) {
	if !IsEnabled() {
		return
	}

	instance.queueUpdate(func(m *uiModel) {
		textStr := ""
		if len(text) > 0 {
			textStr = strings.Join(text, " ")
		}

		if !show {
			m.removeLastLoader()
			return
		}

		m.updateOrAddLoader(textStr)
	})
}

func Snippet(code string) {
	if IsEnabled() {
		instance.queueUpdate(func(m *uiModel) {
			m.content = append(m.content, message{
				text:  code,
				style: snippetStyle,
			})
			trimContent(m)
		})
	}
}

func PackageManager(action, pkg, status string) {
	if !IsEnabled() {
		return
	}

	instance.queueUpdate(func(m *uiModel) {
		m.elements = append(m.elements, Element{
			elementType: TypePackage,
			action:      action,
			packageName: pkg,
			status:      status,
		})
	})
}

func RealtimeMsg(content string) {
	if !IsEnabled() {
		return
	}

	instance.queueUpdate(func(m *uiModel) {
		m.elements = append(m.elements, Element{
			elementType: TypeRealtime,
			content:     content,
		})
	})
}

func Spinner(show bool, text ...string) {
	if !IsEnabled() {
		return
	}

	instance.queueUpdate(func(m *uiModel) {
		textStr := ""
		if len(text) > 0 {
			textStr = strings.Join(text, " ")
		}

		if !show {
			m.removeLastSpinner()
			return
		}

		m.updateOrAddSpinner(textStr)
	})
}

func Stopwatch(start bool, name ...string) {
	if !IsEnabled() {
		return
	}

	instance.queueUpdate(func(m *uiModel) {
		nameStr := ""
		if len(name) > 0 {
			nameStr = strings.Join(name, " ")
		}

		if !start {
			m.removeLastStopwatch()
			return
		}

		m.updateOrAddStopwatch(nameStr)
	})
}

func Table(headers []string, data [][]string) {
	if !IsEnabled() {
		return
	}

	instance.queueUpdate(func(m *uiModel) {
		m.elements = append(m.elements, Element{
			elementType:  TypeTable,
			tableHeaders: headers,
			tableData:    data,
		})
	})
}

func (m *uiModel) removeLastProgress() {
	for i := len(m.elements) - 1; i >= 0; i-- {
		if m.elements[i].elementType == TypeProgress {
			m.elements = append(m.elements[:i], m.elements[i+1:]...)
			return
		}
	}
}

func (m *uiModel) removeLastLoader() {
	for i := len(m.elements) - 1; i >= 0; i-- {
		if m.elements[i].elementType == TypeLoader {
			m.elements = append(m.elements[:i], m.elements[i+1:]...)
			return
		}
	}
}

func (m *uiModel) updateOrAddProgress(percent float64, text string) {
	for i := len(m.elements) - 1; i >= 0; i-- {
		if m.elements[i].elementType == TypeProgress {
			m.elements[i].progress = percent
			m.elements[i].progressText = text
			return
		}
	}

	m.elements = append(m.elements, Element{
		elementType:  TypeProgress,
		progress:     percent,
		progressText: text,
	})
}

func (m *uiModel) updateOrAddLoader(text string) {
	for i := len(m.elements) - 1; i >= 0; i-- {
		if m.elements[i].elementType == TypeLoader {
			m.elements[i].showLoader = true
			m.elements[i].loaderText = text
			return
		}
	}

	m.elements = append(m.elements, Element{
		elementType: TypeLoader,
		showLoader:  true,
		loaderText:  text,
	})
}

func renderPackage(header, body, status string) string {
	actionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	pkgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	statusStyle := lipgloss.NewStyle()

	switch status {
	case "done":
		statusStyle = statusStyle.Foreground(lipgloss.Color("10"))
	case "error":
		statusStyle = statusStyle.Foreground(lipgloss.Color("9")).Bold(true)
	case "working":
		statusStyle = statusStyle.Foreground(lipgloss.Color("214"))
	default:
		statusStyle = statusStyle.Foreground(lipgloss.Color("240"))
	}

	return fmt.Sprintf("%s %s %s",
		actionStyle.Render(header),
		pkgStyle.Render(body),
		statusStyle.Render(status),
	)
}

func renderRealtime(content string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("45")).
		Render("↳ " + content)
}

func renderSpinner(index int, text string) string {
	if text == "" {
		text = "Processing"
	}
	return spinnerStyle.
		Render(fmt.Sprintf("%s %s", spinners[index], text))
}

func renderStopwatch(startTime time.Time, name string) string {
	duration := time.Since(startTime).Round(time.Second)
	timeStr := fmt.Sprintf("%02d:%02d:%02d",
		int(duration.Hours()),
		int(duration.Minutes())%60,
		int(duration.Seconds())%60)

	text := timeStr
	if name != "" {
		text = name + ": " + timeStr
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("51")).
		Render("⏱ " + text)
}

func renderTable(headers []string, data [][]string) string {
	if len(headers) == 0 || len(data) == 0 {
		return ""
	}

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}

	for _, row := range data {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	cellStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

	for i, h := range headers {
		sb.WriteString(headerStyle.Render(fmt.Sprintf("%-*s", colWidths[i]+2, h)))
	}
	sb.WriteString("\n")

	for _, w := range colWidths {
		sb.WriteString(strings.Repeat("─", w+2))
	}
	sb.WriteString("\n")

	for _, row := range data {
		for i, cell := range row {
			sb.WriteString(cellStyle.Render(fmt.Sprintf("%-*s", colWidths[i]+2, cell)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func renderProgressBar(percent float64, text string) string {
	const width = 30
	filled := int(percent / 100 * width)
	unfilled := width - filled

	percentStr := fmt.Sprintf("%3.0f%%", percent)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", unfilled)

	if text == "" {
		text = fmt.Sprintf("%.1f%%", percent)
	}

	percentPart := progressTextStyle.Render(" " + percentStr + " ")
	barPart := progressBarStyle.Render(bar)
	textPart := progressTextStyle.Render(" " + text)

	return textPart + barPart + percentPart
}

func renderLoader(text string) string {
	if text == "" {
		text = "Processing..."
	}
	return loaderStyle.Render("↻ " + text)
}

func renderSnippet(code string) string {
	return snippetStyle.Render(code)
}

func (m *uiModel) removeLastSpinner() {
	for i := len(m.elements) - 1; i >= 0; i-- {
		if m.elements[i].elementType == TypeSpinner {
			m.elements = append(m.elements[:i], m.elements[i+1:]...)
			return
		}
	}
}

func (m *uiModel) updateOrAddSpinner(text string) {
	for i := len(m.elements) - 1; i >= 0; i-- {
		if m.elements[i].elementType == TypeSpinner {
			m.elements[i].showSpinner = true
			m.elements[i].spinnerText = text
			return
		}
	}

	m.elements = append(m.elements, Element{
		elementType: TypeSpinner,
		showSpinner: true,
		spinnerText: text,
	})
}

func (m *uiModel) removeLastStopwatch() {
	for i := len(m.elements) - 1; i >= 0; i-- {
		if m.elements[i].elementType == TypeStopwatch {
			m.elements = append(m.elements[:i], m.elements[i+1:]...)
			return
		}
	}
}

func (m *uiModel) updateOrAddStopwatch(name string) {
	now := time.Now()
	for i := len(m.elements) - 1; i >= 0; i-- {
		if m.elements[i].elementType == TypeStopwatch {
			if name != "" && m.elements[i].content != name {
				continue
			}
			m.elements[i].startTime = now
			m.elements[i].content = name
			return
		}
	}

	m.elements = append(m.elements, Element{
		elementType: TypeStopwatch,
		startTime:   now,
		content:     name,
	})
}
