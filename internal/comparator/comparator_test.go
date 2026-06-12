package comparator

import (
	"testing"

	"github.com/simplesvet/svcomp/pkg/types"
)

func TestCompare(t *testing.T) {
	source := []types.DBObject{
		{Name: "dados", Kind: types.KindTable, Definition: "CREATE TABLE dados (id int primary key)"},
		{Name: "animais", Kind: types.KindTable, Definition: "CREATE TABLE animais (id bigint primary key)"},
		{Name: "v_ativos", Kind: types.KindView, Definition: "CREATE VIEW v_ativos AS SELECT 1"},
	}
	target := []types.DBObject{
		{Name: "animais", Kind: types.KindTable, Definition: "CREATE TABLE animais (id int primary key)"},
		{Name: "pessoas", Kind: types.KindTable, Definition: "CREATE TABLE pessoas (id int primary key)"},
		{Name: "v_ativos", Kind: types.KindView, Definition: "CREATE VIEW v_ativos AS SELECT 1"},
	}

	diffs := New().Compare(source, target)

	if len(diffs) != 3 {
		t.Fatalf("esperava 3 diffs, recebi %d", len(diffs))
	}

	seen := map[string]types.DiffAction{}
	for _, d := range diffs {
		seen[d.Object.Key()] = d.Action
	}

	if seen[types.ObjectKey(types.KindTable, "dados")] != types.ActionCreate {
		t.Fatalf("dados deveria ser create")
	}
	if seen[types.ObjectKey(types.KindTable, "animais")] != types.ActionUpdate {
		t.Fatalf("animais deveria ser update")
	}
	if seen[types.ObjectKey(types.KindTable, "pessoas")] != types.ActionDelete {
		t.Fatalf("pessoas deveria ser delete")
	}
}
