package differ

import (
	"context"
	"strings"
	"testing"

	"github.com/simplesvet/svcomp/pkg/types"
)

func TestGenerateNonTableCreateAndDelete(t *testing.T) {
	gen, err := NewGenerator()
	if err != nil {
		t.Fatalf("erro ao criar generator: %v", err)
	}

	source := map[string]types.DBObject{
		types.ObjectKey(types.KindView, "v_clientes"): {
			Name:       "v_clientes",
			Kind:       types.KindView,
			Definition: "CREATE VIEW v_clientes AS SELECT 1",
		},
	}
	target := map[string]types.DBObject{
		types.ObjectKey(types.KindProcedure, "p_old"): {
			Name:       "p_old",
			Kind:       types.KindProcedure,
			Definition: "CREATE PROCEDURE p_old() SELECT 1",
		},
	}

	diffs := []types.DiffResult{
		{Object: source[types.ObjectKey(types.KindView, "v_clientes")], Action: types.ActionCreate},
		{Object: target[types.ObjectKey(types.KindProcedure, "p_old")], Action: types.ActionDelete},
	}

	sqlText, err := gen.Generate(context.Background(), diffs, source, target)
	if err != nil {
		t.Fatalf("erro ao gerar sql: %v", err)
	}

	if !strings.Contains(sqlText, "DROP VIEW IF EXISTS `v_clientes`;") {
		t.Fatalf("esperava DROP VIEW, recebi: %s", sqlText)
	}
	if !strings.Contains(sqlText, "CREATE VIEW v_clientes AS SELECT 1;") {
		t.Fatalf("esperava CREATE VIEW, recebi: %s", sqlText)
	}
	if !strings.Contains(sqlText, "DROP PROCEDURE IF EXISTS `p_old`;") {
		t.Fatalf("esperava DROP PROCEDURE, recebi: %s", sqlText)
	}
}
