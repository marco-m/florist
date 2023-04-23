package florist_test

import (
	"bytes"
	"os"
	"os/user"
	"path"
	"testing"
	"testing/fstest"
	"text/template"

	"gotest.tools/v3/assert"

	"github.com/marco-m/florist/pkg/florist"
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

func TestCopyTemplateFromFs(t *testing.T) {
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
		tmpDir := t.TempDir()
		srcPath := "fruits.txt.tpl"
		dstPath := path.Join(tmpDir, "fruits.txt")
		fruitBox := &FruitBox{Amount: 42, Fruit: "bananas"}
		currUser, err := user.Current()
		assert.NilError(t, err)
		fsys := fstest.MapFS{
			srcPath: &fstest.MapFile{Data: []byte(tc.tplContents)},
		}

		err = florist.CopyTemplateFs(fsys, srcPath, dstPath, 0640, currUser.Username,
			fruitBox, tc.sepL, tc.sepR)
		assert.NilError(t, err)

		rendered, err := os.ReadFile(dstPath)
		assert.NilError(t, err)
		have := string(rendered)
		assert.Equal(t, have, tc.want)
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
