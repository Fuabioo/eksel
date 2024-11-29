package main

import (
	"bytes"
	"flag"
	"fmt"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
	"github.com/xuri/excelize/v2"
)

const maxLen = 999

var (
	debug           bool
	performanceMode bool
	inputFilename   string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.BoolVar(&performanceMode, "performance", false, "Enable performance mode for large files")
	flag.Parse()

	// input file name is the first argument
	if flag.NArg() > 0 {
		inputFilename = flag.Arg(0)
	}

	if inputFilename == "" {
		log.Fatal("Please provide a filename")
	}
}

var (
	file   *excelize.File
	sheets []string
)

func main() {

	var buffer bytes.Buffer
	log.SetOutput(&buffer)
	log.SetColorProfile(termenv.TrueColor)
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	defer func() {
		fmt.Print(buffer.String())
	}()

	// use excelize to read the file and get the data

	var err error

	file, err = excelize.OpenFile(inputFilename)
	if err != nil {
		log.Error(err)
		return
	}
	defer file.Close()

	// read all the sheets
	sheets = file.GetSheetList()
	log.Debug("Read file sheets", "sheets", sheets)

	if len(sheets) == 0 {
		log.Error("No sheets found in the file")
		return
	}

	tables, err := loadSheets(
		table.WithStyles(tableStylesFocused),
	)
	if err != nil {
		log.Error(err)
		return
	}

	var p *tea.Program

	model := model{
		tables: tables,
		paginator: paginator.New(
			paginator.WithTotalPages(len(sheets)),
		),
	}

	p = tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		log.Error(err)
		return
	}

}

func loadSheets(opts ...table.Option) ([]table.Model, error) {

	result := make([]table.Model, 0, len(sheets))

	for i, sheet := range sheets {

		// use a streamed approach to read the rows

		columns := make([]table.Column, 0)
		rows := make([]table.Row, 0)

		excelRows, err := file.Rows(sheet)
		if err != nil {
			return nil, err
		}

		header := true
		index := 0
		for excelRows.Next() {

			row, err := excelRows.Columns()
			if err != nil {
				return nil, err
			}

			if header {

				for _, column := range row {
					length := len(column)
					if length > maxLen {
						length = maxLen
					}
					columns = append(columns, table.Column{
						Title: column,
						Width: length,
					})
				}

				log.Debug("Read columns", "len", len(columns))
				header = false
				continue
			}

			row = append([]string{fmt.Sprint(index)}, row...)
			index++

			if !performanceMode {
				for i, column := range columns {
					if i >= len(row) {
						break
					}
					rowValue := row[i]
					rowLength := len(rowValue)
					if rowLength > column.Width {
						if rowLength > maxLen {
							rowLength = maxLen
						}
						columns[i].Width = rowLength
					}
				}
			}

			rows = append(rows, table.Row(row))
		}

		amount := fmt.Sprint(len(rows))

		columns = append([]table.Column{
			{
				Title: "",
				Width: len(amount),
			},
		}, columns...)

		opts := append(opts,
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(i == 0),
		)

		result = append(result, table.New(opts...))

	}

	return result, nil
}
