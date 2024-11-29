package main

import "github.com/charmbracelet/lipgloss"

// Base Colors
var (
	Primary         = lipgloss.Color("#00bbf9") // Accent color
	Secondary       = lipgloss.Color("#9b5de5") // Secondary accent
	Highlight       = lipgloss.Color("#fee440") // Highlight color
	Inactive        = lipgloss.Color("#666")    // Inactive/not focused elements
	OpacityReduced  = lipgloss.Color("#333")    // Reduced opacity
	OpacityReduced2 = lipgloss.Color("#323232") // Reduced opacity
	OpacityReduced3 = lipgloss.Color("#353535") // Reduced opacity
)

// Light Theme
var (
	BackgroundLight = lipgloss.Color("#ffffff")
	ForegroundLight = lipgloss.Color("#222")
)

// Dark Theme
var (
	BackgroundDark = lipgloss.Color("#222")
	ForegroundDark = lipgloss.Color("#ddd")
)
