package htmlViewer

import (
	_ "embed"
	"html/template"
	"io"
	"os"
)

//go:embed template.html
var templateString string

func FillTemplate(instructions []int, w io.Writer) error {
	tmpl, err := template.New("instructions").Parse(templateString)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, instructions)
}

func WriteInstructions(filename string, instructions []int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return FillTemplate(instructions, file)
}
