package florist

import (
	"bytes"
	"testing"
	"testing/fstest"
	"text/template"

	"github.com/marco-m/rosina/assert"
)

func TestUnderstandTemplate(t *testing.T) {
	type Inventory struct {
		Material string
		Count    uint
	}
	sweaters := Inventory{Material: "wool", Count: 17}
	tmpl, err := template.New("name").Parse("{{.Count}} items are made of {{.Material}}")
	assert.NoError(t, err, "template.New")

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, sweaters)
	assert.NoError(t, err, "template.Execute")

	assert.Equal(t, buf.String(), "17 items are made of wool", "rendered")
}

func TestUnderstandTemplateFailure(t *testing.T) {
	type Inventory struct {
		Material string
		Count    uint
	}
	sweaters := Inventory{Material: "wool", Count: 17}
	tmpl, err := template.New("name").Parse("{{.Banana}} items are made of {{.Material}}")
	assert.NoError(t, err, "template.New")

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, sweaters)
	assert.ErrorContains(t, err,
		`template: name:1:2: executing "name" at <.Banana>: can't evaluate field Banana in type florist.Inventory`)
}

func TestRenderTemplate(t *testing.T) {
	type FruitBox struct {
		Amount int
		Fruit  string
	}
	type testCase struct {
		name        string
		tplContents string
		sepL, sepR  string
		want        string
	}

	run := func(t *testing.T, tc testCase) {
		srcPath := "fruits.txt.tpl"
		fruitBox := &FruitBox{Amount: 42, Fruit: "bananas"}
		fsys := fstest.MapFS{
			srcPath: &fstest.MapFile{Data: []byte(tc.tplContents)},
		}

		rendered, err := renderTemplate(fsys, srcPath, fruitBox, tc.sepL, tc.sepR)

		assert.NoError(t, err, "renderTemplate")
		assert.Equal(t, rendered, tc.want, "rendered")
	}

	testCases := []testCase{
		{
			name:        "default separators",
			tplContents: "{{.Amount}} kg of {{.Fruit}}",
			sepL:        "",
			sepR:        "",
			want:        "42 kg of bananas",
		},
		{
			name:        "custom separators",
			tplContents: "<<.Amount>> kg of <<.Fruit>>",
			sepL:        "<<",
			sepR:        ">>",
			want:        "42 kg of bananas",
		},
		{
			name:        "custom separators, preserves default",
			tplContents: "<<.Amount>> kg {{ of <<.Fruit>>",
			sepL:        "<<",
			sepR:        ">>",
			want:        "42 kg {{ of bananas",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
