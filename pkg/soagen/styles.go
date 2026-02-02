// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package soagen

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// getTextStyle returns the standard text style for string fields
func getTextStyle() *excelize.Style {
	return &excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
	}
}

// getHeaderStyle returns the standard header style
func getHeaderStyle() *excelize.Style {
	return &excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10, Color: "#000000"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D9D9D9"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	}
}

// getCellStyle returns the standard cell style
func getCellStyle() *excelize.Style {
	return &excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	}
}

func createCellStyle(f *excelize.File, style *excelize.Style) (int, error) {
	styleID, err := f.NewStyle(style)
	if err != nil {
		return 0, fmt.Errorf("cannot create style: %w", err)
	}
	return styleID, nil
}

func applyDataValidation(f *excelize.File, sheetName, col string, row int, validation []string) error {
	if len(validation) == 0 {
		return nil
	}

	dv := excelize.NewDataValidation(true)
	dv.Sqref = fmt.Sprintf("%s%d:%s1000", col, row, col) // Apply to reasonable range
	_ = dv.SetDropList(validation)
	dv.SetError(excelize.DataValidationErrorStyleStop, "Invalid Input", "Please select from the dropdown list.")
	if err := f.AddDataValidation(sheetName, dv); err != nil {
		return fmt.Errorf("cannot add data validation: %w", err)
	}

	return nil
}

func setColumnWidth(f *excelize.File, sheetName, col string, width float64) error {
	if width > 0 {
		if err := f.SetColWidth(sheetName, col, col, width); err != nil {
			return fmt.Errorf("cannot set column width: %w", err)
		}
	}
	return nil
}
