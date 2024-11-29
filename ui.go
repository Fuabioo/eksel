package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type model struct {
	width            int
	height           int
	horizontalCursor int
	tables           []table.Model
	paginator        paginator.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	currentIndex := m.paginator.Page
	paginatorFocused := !m.tables[currentIndex].Focused()
	horizontalEmitted := false

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":

			if m.tables[currentIndex].Focused() {
				m.tables[currentIndex].Blur()
			} else {
				m.tables[currentIndex].Focus()
			}
			t, cmd := m.tables[currentIndex].Update(msg)
			cmds = append(cmds, cmd)

			m.tables[currentIndex] = t

		case "left":
			horizontalEmitted = true
			if !paginatorFocused {

				m.horizontalCursor--
				if m.horizontalCursor < 0 {
					m.horizontalCursor = 0
				}
			} else {
				m.horizontalCursor = 0
			}

		case "right":
			horizontalEmitted = true
			if !paginatorFocused {
				m.horizontalCursor++
				if m.horizontalCursor >= len(m.tables[currentIndex].Columns()) {
					m.horizontalCursor = len(m.tables[currentIndex].Columns()) - 1
				}
			} else {
				m.horizontalCursor = 0
			}

		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.tables[currentIndex].SelectedRow()[1]),
			)

		}

	case tea.WindowSizeMsg:

		const maxHeight = 666

		m.width = msg.Width
		m.height = msg.Height

		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		paddings := 0
		height := msg.Height - verticalMarginHeight - paddings
		if height > maxHeight {
			height = maxHeight
		}

		if msg.Height < verticalMarginHeight {

			log.Error("Terminal height is too small")

			return m, tea.Quit
		}

		// update the height of the table
		for index, t := range m.tables {
			t.SetHeight(height)
			m.tables[index] = t
		}
	}

	for index, t := range m.tables {
		t, cmd = t.Update(msg)
		m.tables[index] = t
		cmds = append(cmds, cmd)
	}

	if !horizontalEmitted || paginatorFocused {
		m.paginator, cmd = m.paginator.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

var (
	tableStylesFocused = table.DefaultStyles()
	tableStylesBlurred = table.DefaultStyles()
)

func init() {
	tableStylesFocused.Selected = tableStylesFocused.Selected.Foreground(Primary)
	tableStylesFocused.Header = tableStylesFocused.Header.Foreground(Secondary)

	// tone down al little bit the colors for the blurred table
	tableStylesBlurred.Selected = tableStylesBlurred.Selected.Foreground(Inactive)
	tableStylesBlurred.Header = tableStylesBlurred.Header.Foreground(Inactive)
	tableStylesBlurred.Cell = tableStylesBlurred.Cell.Foreground(Inactive)
}

func (m model) View() string {
	currentIndex := m.paginator.Page
	currentActiveTable := m.tables[currentIndex]

	styles := tableStylesBlurred
	if currentActiveTable.Focused() {
		styles = tableStylesFocused
	}

	newTable := table.New(
		table.WithFocused(currentActiveTable.Focused()),
		table.WithHeight(currentActiveTable.Height()),
		table.WithColumns(currentActiveTable.Columns()),
		table.WithRows(currentActiveTable.Rows()),
		table.WithStyles(styles),
	)
	newTable.SetCursor(currentActiveTable.Cursor())

	// use the horizontalCursor to determine horizontal scroll
	// for the table since the table does not support horizontal scrolling

	if m.horizontalCursor > 0 {

		// copy the table and remove the left columns until the horizontalCursor
		// is reached
		currentRows := currentActiveTable.Rows()
		newRows := make([]table.Row, len(currentRows))
		for i, row := range currentRows {
			if len(row) <= m.horizontalCursor {
				newRows[i] = []string{}
				continue
			}
			newRows[i] = row[m.horizontalCursor:]
		}
		newTable.SetRows(newRows)

		currentCols := currentActiveTable.Columns()
		currentCols = currentCols[m.horizontalCursor:]
		newTable.SetColumns(currentCols)
	}

	footer := m.footerView()
	footerHeight := lipgloss.Height(footer)

	content := lipgloss.PlaceVertical(m.height-footerHeight, lipgloss.Top,
		lipgloss.NewStyle().Margin(1, 0).Render(newTable.View()),
	)

	return lipgloss.JoinVertical(lipgloss.Top,
		content,
		footer,
	)
}

func (m model) headerView() string {
	return ""
}

var paginatorBaseStyle = lipgloss.NewStyle().
	Padding(0, 1).
	Bold(true).
	Background(OpacityReduced)

func (m model) footerView() string {

	paginationFocused := !m.tables[m.paginator.Page].Focused()

	m.paginator.ActiveDot = paginatorBaseStyle.
		Render("•")
	m.paginator.InactiveDot = paginatorBaseStyle.
		Bold(false).
		Render("•")
	if paginationFocused {
		m.paginator.ActiveDot = paginatorBaseStyle.
			Foreground(Primary).
			Render("•")
		m.paginator.InactiveDot = paginatorBaseStyle.
			Foreground(Inactive).
			Bold(false).
			Render("•")
	}

	currentMode := t("mode.navigation.table")
	currentModeStyle := lipgloss.NewStyle().
		Background(Secondary).
		Transform(strings.ToUpper).
		Padding(0, 1).
		Bold(true)
	if paginationFocused {
		currentMode = t("mode.navigation.sheet")
		currentModeStyle = currentModeStyle.
			Background(Primary)
	}

	var builder strings.Builder

	builder.WriteString(currentModeStyle.Render(currentMode))

	middlePartStyle := lipgloss.NewStyle().
		Background(OpacityReduced2)

	builder.WriteString(middlePartStyle.Render(" "))
	builder.WriteString(middlePartStyle.Render(t("context.filename", inputFilename)))
	builder.WriteString(middlePartStyle.Render(" ↪ "))
	builder.WriteString(middlePartStyle.Render(t("context.sheetname", sheets[m.paginator.Page])))
	builder.WriteString(middlePartStyle.Render(" "))

	paginator := middlePartStyle.Padding(0, 1).Render(
		t("context.paginator", m.paginator.Page+1, len(m.tables)),
	)

	return lipgloss.JoinHorizontal(lipgloss.Right,
		lipgloss.PlaceHorizontal(m.width-lipgloss.Width(paginator), lipgloss.Left, builder.String(),
			lipgloss.WithWhitespaceBackground(OpacityReduced),
		),
		paginator,
	)
}
