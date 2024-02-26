package provisioner_test

import (
	"testing"

	"github.com/go-quicktest/qt"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/provisioner"
)

var _ florist.Flower = (*TestFlower)(nil)

type TestFlower struct {
	Inst
	Conf
}

type Inst struct {
	name string
	desc string
}

type Conf struct{}

func (fl *TestFlower) String() string {
	return fl.name
}

func (fl *TestFlower) Description() string {
	return fl.desc
}

func (fl *TestFlower) Init() error {
	return nil
}

func (fl *TestFlower) Embedded() []string {
	return nil
}

func (fl *TestFlower) Install() error {
	return nil
}

func (fl *TestFlower) Configure() error {
	return nil
}

func TestProvisionerAddFlowersSuccess(t *testing.T) {
	prov, err := provisioner.New(florist.CacheValidity)
	qt.Assert(t, qt.IsNil(err))

	err = prov.AddFlowers(
		&TestFlower{Inst: Inst{name: "A", desc: "desc A"}},
		&TestFlower{Inst: Inst{name: "B", desc: "desc B"}},
	)
	qt.Assert(t, qt.IsNil(err))

	have := prov.Flowers()
	qt.Assert(t, qt.Equals(have["A"].String(), "A"))
	qt.Assert(t, qt.Equals(have["B"].String(), "B"))
	qt.Assert(t, qt.Equals(len(have), 2))
}

func TestProvisionerAddFlowersFailure(t *testing.T) {
	type testCase struct {
		name    string
		flowers []florist.Flower
		want    string
	}

	test := func(t *testing.T, tc testCase) {
		prov, err := provisioner.New(florist.CacheValidity)
		qt.Assert(t, qt.IsNil(err))

		err = prov.AddFlowers(tc.flowers...)
		qt.Assert(t, qt.ErrorMatches(err, tc.want))
	}

	testCases := []testCase{
		{
			name:    "empty name",
			flowers: []florist.Flower{&TestFlower{}},
			want:    `Provisioner\.AddFlowers: flower 0 has empty name`,
		},
		{
			name:    "empty description",
			flowers: []florist.Flower{&TestFlower{Inst: Inst{name: "A"}}},
			want:    `Provisioner\.AddFlowers: flower A has empty description`,
		},
		{
			name: "same name",
			flowers: []florist.Flower{
				&TestFlower{Inst: Inst{name: "A", desc: "X"}},
				&TestFlower{Inst: Inst{name: "A", desc: "Y"}},
			},
			want: `Provisioner\.AddFlowers: flower with same name already exists: A`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}
