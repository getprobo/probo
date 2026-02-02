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
	"strings"

	"github.com/xuri/excelize/v2"
)

func GenerateSOAExcel(data SOAData) ([]byte, error) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheetName := "State of Applicability"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("cannot create sheet: %w", err)
	}

	if err := f.DeleteSheet("Sheet1"); err != nil {
		return nil, fmt.Errorf("cannot delete sheet: %w", err)
	}

	if err := setupSOAHeader(f, sheetName); err != nil {
		return nil, fmt.Errorf("cannot setup Excel header: %w", err)
	}

	if err := populateSOAData(f, sheetName, data.Rows); err != nil {
		return nil, fmt.Errorf("cannot populate Excel data: %w", err)
	}

	if err := applySOAFinalFormatting(f, sheetName, len(data.Rows)); err != nil {
		return nil, fmt.Errorf("cannot apply final formatting: %w", err)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("cannot write Excel file to buffer: %w", err)
	}

	return buf.Bytes(), nil
}

func setupSOAHeader(f *excelize.File, sheetName string) error {
	headerStyle, err := f.NewStyle(getHeaderStyle())
	if err != nil {
		return err
	}

	cellStyle, err := f.NewStyle(getCellStyle())
	if err != nil {
		return err
	}

	return applyHeaderLayout(f, sheetName, getSOAHeaderLayout(), headerStyle, cellStyle)
}

func populateSOAData(f *excelize.File, sheetName string, rows []SOARowData) error {
	dataStartRow := 8

	for i, rowData := range rows {
		isFirstRow := i == 0
		filterColumns, err := writeSOARow(f, sheetName, dataStartRow+i, rowData, isFirstRow)
		if err != nil {
			return fmt.Errorf("cannot write row %d: %w", dataStartRow+i, err)
		}

		if isFirstRow && len(filterColumns) > 0 {
			if err := applySOAAutoFilter(f, sheetName, filterColumns); err != nil {
				return fmt.Errorf("cannot apply auto filter: %w", err)
			}
		}
	}

	return nil
}

func writeSOARow(f *excelize.File, sheetName string, row int, data SOARowData, isFirstRow bool) ([]string, error) {
	var allFilterColumns []string

	fields := []struct {
		name  string
		value interface{}
	}{
		{"SectionTitle", data.SectionTitle},
		{"ControlName", data.ControlName},
		{"Applicability", data.Applicability},
		{"Regulatory", data.Regulatory},
		{"Contractual", data.Contractual},
		{"BestPractice", data.BestPractice},
		{"RiskAssessment", data.RiskAssessment},
		{"ExclusionJustification", data.ExclusionJustification},
		{"SecurityMeasures", data.SecurityMeasures},
	}

	for _, field := range fields {
		fieldConfig := getSOAFieldConfig(field.name)
		if fieldConfig == nil {
			continue
		}

		filterColumns, err := processField(f, sheetName, row, field.value, *fieldConfig, isFirstRow)
		if err != nil {
			return nil, fmt.Errorf("cannot write field %s: %w", field.name, err)
		}

		allFilterColumns = append(allFilterColumns, filterColumns...)
	}

	return allFilterColumns, nil
}

func getSOAFieldConfig(fieldName string) *FieldConfiguration {
	for _, config := range soaFieldConfigs {
		if config.Field == fieldName {
			return &config
		}
	}
	return nil
}

func applySOAAutoFilter(f *excelize.File, sheetName string, filterColumns []string) error {
	if len(filterColumns) == 0 {
		return nil
	}

	firstCol, lastCol := filterColumns[0], filterColumns[0]
	for _, col := range filterColumns {
		if col < firstCol {
			firstCol = col
		}
		if col > lastCol {
			lastCol = col
		}
	}

	filterRange := fmt.Sprintf("%s7:%s1000", firstCol, lastCol)
	return f.AutoFilter(sheetName, filterRange, []excelize.AutoFilterOptions{})
}

func applySOAFinalFormatting(f *excelize.File, sheetName string, dataRowCount int) error {
	if dataRowCount > 0 {
		headerRowHeight := 30.0
		_ = f.SetRowHeight(sheetName, 6, headerRowHeight)
		_ = f.SetRowHeight(sheetName, 7, headerRowHeight)

		return f.SetPanes(sheetName, &excelize.Panes{
			Freeze:      true,
			XSplit:      0,
			YSplit:      7,
			TopLeftCell: "A8",
		})
	}
	return nil
}

func processField(f *excelize.File, sheetName string, row int, value interface{}, config FieldConfiguration, isFirstRow bool) ([]string, error) {
	switch v := value.(type) {
	case ExcelMarshaler:
		marshaler := v
		excelValue := marshaler.MarshalExcel()

		styleID, err := createCellStyle(f, excelValue.Style)
		if err != nil {
			return nil, err
		}

		return writeSingleColumnField(f, sheetName, row, excelValue, &config, styleID, isFirstRow)

	case string:
		if len(config.Columns) == 0 {
			return []string{}, nil
		}

		col := config.Columns[0]
		cellRef := fmt.Sprintf("%s%d", col, row)

		if err := f.SetCellValue(sheetName, cellRef, v); err != nil {
			return nil, fmt.Errorf("cannot set cell value: %w", err)
		}

		textStyleID, err := createCellStyle(f, getTextStyle())
		if err != nil {
			return nil, fmt.Errorf("cannot create text style: %w", err)
		}
		if err := f.SetCellStyle(sheetName, cellRef, cellRef, textStyleID); err != nil {
			return nil, fmt.Errorf("cannot set cell style: %w", err)
		}

		if isFirstRow {
			width := config.DefaultWidth
			if len(config.Width) > 0 {
				width = config.Width[0]
			}
			if err := f.SetColWidth(sheetName, col, col, width); err != nil {
				return nil, fmt.Errorf("cannot set column width: %w", err)
			}
		}

		if config.HasFilter {
			if len(config.FilterColumns) > 0 {
				return config.FilterColumns, nil
			}
			return []string{col}, nil
		}
		return []string{}, nil

	case []string:
		var formattedLines []string
		for _, item := range v {
			lines := strings.Split(item, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					formattedLines = append(formattedLines, "• "+strings.TrimSpace(line))
				}
			}
		}
		joinedValue := strings.Join(formattedLines, "\n")

		if len(config.Columns) == 0 {
			return []string{}, nil
		}

		col := config.Columns[0]
		cellRef := fmt.Sprintf("%s%d", col, row)

		if err := f.SetCellValue(sheetName, cellRef, joinedValue); err != nil {
			return nil, fmt.Errorf("cannot set cell value: %w", err)
		}

		textStyleID, err := createCellStyle(f, getTextStyle())
		if err != nil {
			return nil, fmt.Errorf("cannot create text style: %w", err)
		}
		if err := f.SetCellStyle(sheetName, cellRef, cellRef, textStyleID); err != nil {
			return nil, fmt.Errorf("cannot set cell style: %w", err)
		}

		if isFirstRow {
			width := config.DefaultWidth
			if len(config.Width) > 0 {
				width = config.Width[0]
			}
			if err := f.SetColWidth(sheetName, col, col, width); err != nil {
				return nil, fmt.Errorf("cannot set column width: %w", err)
			}
		}

		if config.HasFilter {
			if len(config.FilterColumns) > 0 {
				return config.FilterColumns, nil
			}
			return []string{col}, nil
		}
		return []string{}, nil

	case *bool:
		yesNoValue := boolToYesNo(v)
		excelValue := yesNoValue.MarshalExcel()

		styleID, err := createCellStyle(f, excelValue.Style)
		if err != nil {
			return nil, err
		}

		return writeSingleColumnField(f, sheetName, row, excelValue, &config, styleID, isFirstRow)

	default:
		return nil, fmt.Errorf("no handler found for field type: %T", value)
	}
}

func GenerateExcel(data SOAData) ([]byte, error) {
	return GenerateSOAExcel(data)
}

func writeSingleColumnField(f *excelize.File, sheetName string, row int, excelValue ExcelValue, colDef *FieldConfiguration, styleID int, isFirstRow bool) ([]string, error) {
	if len(colDef.Columns) == 0 {
		return []string{}, nil
	}

	col := colDef.Columns[0]
	cellRef := fmt.Sprintf("%s%d", col, row)

	_ = f.SetCellValue(sheetName, cellRef, excelValue.Value)
	_ = f.SetCellStyle(sheetName, cellRef, cellRef, styleID)

	if isFirstRow {
		if err := applyDataValidation(f, sheetName, col, row, excelValue.Validation); err != nil {
			return nil, err
		}

		if err := setColumnWidth(f, sheetName, col, excelValue.Width); err != nil {
			return nil, err
		}
	}

	if colDef.HasFilter {
		if len(colDef.FilterColumns) > 0 {
			return colDef.FilterColumns, nil
		}
		return []string{col}, nil
	}
	return []string{}, nil
}
