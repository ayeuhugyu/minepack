package util

import (
	"minepack/core/project"

	"github.com/charmbracelet/lipgloss/v2"
)

var contentDataStyle = lipgloss.NewStyle().
	Padding(0, 1).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Margin(0, 1)

var boldStyle = lipgloss.NewStyle().Bold(true)
var grayStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
var curseforgeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ef7d4fff"))
var modrinthStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#48cf7aff"))

func joinLinesByNewline(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

func stylizeDependencyType(depType string) string {
	switch depType {
	case "required":
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4487aeff")).Render(depType)
	case "optional":
		return grayStyle.Render(depType)
	case "incompatible":
		return FormatError(depType)
	case "embedded":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#c1b04fff")).Render(depType)
	}
	return depType
}

func FormatContentData(data project.ContentData) string {
	var lines []string
	lines = append(lines, boldStyle.Render(data.Name)+grayStyle.Render(" ("+data.Slug+")")) // make the slug gray
	switch data.Source {
	case project.Modrinth:
		lines = append(lines, modrinthStyle.Render(data.PageUrl))
	case project.Curseforge:
		lines = append(lines, curseforgeStyle.Render(data.PageUrl))
	default:
		lines = append(lines, grayStyle.Render("(no page)"))
	}
	if len(data.Dependencies) > 0 {
		lines = append(lines, "\ndependencies:")
		for _, dep := range data.Dependencies {
			lines = append(lines, " - "+stylizeDependencyType(project.DependencyTypeToString(dep.DependencyType))+": "+dep.Name+grayStyle.Render(" ("+dep.Slug+")"))
		}
	}
	return contentDataStyle.Render(joinLinesByNewline(lines))
}
