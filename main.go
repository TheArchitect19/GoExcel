package main

import (
	"fmt"
	"log"
	"github.com/tealeg/xlsx"
	"github.com/gofiber/fiber/v2"
	
)

func main() {
	app := fiber.New(fiber.Config{
		BodyLimit:   1024 * 1024 * 1024,
		Concurrency: 256 * 1024,
	})
	app.Get("/", alive)
	app.Post("/generate", createExcelHandler);
	// app.Post("/validate", validator)
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
		dataType := c.FormValue(key)
		for _, value := range values {
			if isValidType(dataType, value) {
				entry := map[string]string{
					key: value,
				}
				data = append(data, entry)
			}
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
func isValidType(fieldType, values string) bool {
	switch fieldType {
	case "string":
		return true
	case "email":
		return true
	case "phone":
		return true
	default:
		return false
	}
}