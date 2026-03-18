package ui

import "charm.land/lipgloss/v2"

var (
	orange       = lipgloss.Color("#FFA500")
	bannerStyle  = lipgloss.NewStyle().Foreground(orange).Bold(true)
	infoStyle    = lipgloss.NewStyle()
	successStyle = lipgloss.NewStyle().Bold(true)
	errorStyle   = lipgloss.NewStyle().Bold(true)
)

func RenderBanner(message string) string {
	return bannerStyle.Render(message)
}

func RenderInfo(message string) string {
	return infoStyle.Render(message)
}

func RenderSuccess(message string) string {
	return successStyle.Render(message)
}

func RenderError(message string) string {
	return errorStyle.Render("error: " + message)
}
