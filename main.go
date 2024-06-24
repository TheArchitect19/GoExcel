package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/tealeg/xlsx"
	"github.com/xuri/excelize/v2"
)





type ValidationRule struct {
	ColumnName string `json:"column_name"`
	DataType   string `json:"data_type"`
}

type Warning struct {
	CellAddress string `json:"cell_address"`
	Message     string `json:"message"`
}


func main() {
	app := fiber.New(fiber.Config{
		BodyLimit:   1024 * 1024 * 1024,
		Concurrency: 256 * 1024,
	})
	app.Get("/", alive)
	app.Post("/generate", createExcelHandler);
	app.Post("/validate", validator)
	fmt.Println("Server started at :8080")
	log.Fatal(app.Listen(":8080"))
}

//function to check the server is alive

func alive(c *fiber.Ctx) error {
	return c.SendString("Golang is alive")
}


func createExcelHandler(c *fiber.Ctx) error {
    form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Failed to parse the form data")
	}
	// fmt.Println(form.Value)
	var data []map[string]string

	for key, values := range form.Value {
		// dataType := c.FormValue(key)
		for _, value := range values {
			entry := map[string]string{
				key: value,
			}
			data = append(data, entry)
		}
	}
    file := xlsx.NewFile()
    sheet, err := file.AddSheet("Data")
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Failed to create Excel sheet")
    }

    headerRow := sheet.AddRow()
    for _, entry := range data {
        for key := range entry {
            cell := headerRow.AddCell()
            cell.Value = key
        }
    }

    filename := "output.xlsx"
    if err := file.Save(filename); err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Failed to save Excel file")
    }
    return c.Download(filename, filename)
}



func validator(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Failed to parse form")
	}

	fileHeader := form.File["file"][0]
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to open file")
	}
	defer file.Close()

	// Open Excel file
	f, err := excelize.OpenReader(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to read Excel file")
	}

	rules := make(map[string]string)
	for key, values := range form.Value {
		if len(values) > 0 {
			rules[key] = values[0]
		}
	}
	fmt.Println("Validation rules:", rules)

	sheetName := f.GetSheetName(0)
	fmt.Println("Sheet name:", sheetName)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to get rows")
	}
	fmt.Println("Rows:", rows)

	headers := rows[0]
	fmt.Println("Headers:", headers)
	var warnings []Warning
	for i, row := range rows[1:] {
		fmt.Println("Row:", row)
		for columnName, dataType := range rules {
			colIndex := getColumnIndex(headers, columnName)
			fmt.Println(colIndex)
			if colIndex == -1 {
				continue
			}
			cell := ""
			if colIndex < len(row) {
				cell = row[colIndex]
			}
			fmt.Printf("Validating cell at column '%s' with value '%s'\n", columnName, cell)
			if !validateCell(cell, dataType) {
				cellAddress, _ := excelize.CoordinatesToCellName(colIndex+1, i+2)
				warnings = append(warnings, Warning{
					CellAddress: cellAddress,
					Message:     "Invalid data type for column " + columnName,
				})
			}
		}
	}


	return c.JSON(warnings)
}

func getColumnIndex(headers []string, columnName string) int {
	for i, header := range headers {
		if header == columnName {
			return i
		}
	}
	return -1
}

func validateCell(cell, dataType string) bool {
	switch dataType {
		case "string":
			return true
		case "email":
			return strings.Contains(cell, "@") && cell == strings.ToLower(cell)
		case "number":
			_, err := strconv.Atoi(cell)
			return err == nil
		case "name":
			match, _ := regexp.MatchString(`^[a-zA-Z]+$`, cell)
			return match
		case "address":
			match, _ := regexp.MatchString(`^[a-zA-Z0-9\s,.'-]+$`, cell)
			return match
		case "phone":
			match, _ := regexp.MatchString(`^[0-9]{10}$`, cell)
			return match
		default:
			return false
	}
}