package florist_test

import (
	"bytes"
	"os"
	"os/user"
	"path"
	"testing"
	"text/template"

	"github.com/marco-m/florist"
	"gotest.tools/v3/assert"
)

func TestUnderstandTemplate(t *testing.T) {
	type Inventory struct {
		Material string
		Count    uint
	}
	sweaters := Inventory{Material: "wool", Count: 17}
	tmpl, err := template.New("name").Parse("{{.Count}} items are made of {{.Material}}")
	assert.NilError(t, err)

	var buf bytes.Buffer
	assert.NilError(t, tmpl.Execute(&buf, sweaters))

	assert.Equal(t, buf.String(), "17 items are made of wool")
}

func TestUnderstandTemplateFailure(t *testing.T) {
	type Inventory struct {
		Material string
		Count    uint
	}
	sweaters := Inventory{Material: "wool", Count: 17}
	tmpl, err := template.New("name").Parse("{{.Banana}} items are made of {{.Material}}")
	assert.NilError(t, err)

	var buf bytes.Buffer
	assert.Error(t, tmpl.Execute(&buf, sweaters),
		"template: name:1:2: executing \"name\" at <.Banana>: can't evaluate field Banana in type florist_test.Inventory")
}

func TestCopyFileTemplate(t *testing.T) {
	type FruitBox struct {
		Amount int
		Fruit  string
	}
	tmpDir := t.TempDir()
	tmplFilePath := path.Join(tmpDir, "fruits.txt.tmpl")
	renderedFilePath := path.Join(tmpDir, "fruits.txt")
	tmplContents := "{{.Amount}} kg of {{.Fruit}}"
	assert.NilError(t, os.WriteFile(tmplFilePath, []byte(tmplContents), 0640))
	box := &FruitBox{Amount: 42, Fruit: "bananas"}
	currUser, err := user.Current()
	assert.NilError(t, err)

	assert.NilError(t, florist.CopyFileTemplate(tmplFilePath, renderedFilePath, 0640, currUser, box))

	renderedContents, err := os.ReadFile(renderedFilePath)
	assert.NilError(t, err)
	assert.Equal(t, string(renderedContents), "42 kg of bananas")
}
