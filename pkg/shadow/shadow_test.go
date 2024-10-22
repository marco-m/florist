package shadow_test

import (
	"testing"
	"time"

	"github.com/go-quicktest/qt"

	"github.com/marco-m/florist/pkg/shadow"
)

func TestExplore1(t *testing.T) {
	db, err := shadow.OpenFiles("testdata/passwd", "testdata/shadow", shadow.ReadOnly)
	qt.Assert(t, qt.IsNil(err))

	u, found := db.Lookup("root")
	qt.Assert(t, qt.IsTrue(found))

	qt.Assert(t, qt.Equals(u.Home, "/root"))
	want, err := time.Parse("2006-01-02", "2024-01-13")
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(u.PassLastChanged, want))

	err := db.Write(u)
	qt.Assert(t, qt.ErrorIs(err, shadow.ErrOpenReadOnly))
}
