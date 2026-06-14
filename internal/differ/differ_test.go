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
		types.ObjectKey(types.KindTable, "clientes"): {
			Name:       "clientes",
			Kind:       types.KindTable,
			Definition: "CREATE TABLE clientes (id INT)",
		},
		types.ObjectKey(types.KindTable, "empresa"): {
			Name:       "empresa",
			Kind:       types.KindTable,
			Definition: "CREATE TABLE empresa (id INT, criado_em TIMESTAMP)",
		},
	}
	target := map[string]types.DBObject{
		types.ObjectKey(types.KindProcedure, "p_old"): {
			Name:       "p_old",
			Kind:       types.KindProcedure,
			Definition: "CREATE PROCEDURE p_old() SELECT 1",
		},
		types.ObjectKey(types.KindTable, "empresa"): {
			Name:       "empresa",
			Kind:       types.KindTable,
			Definition: "CREATE TABLE empresa (id INT)",
		},
	}

	diffs := []types.DiffResult{
		{Object: source[types.ObjectKey(types.KindView, "v_clientes")], Action: types.ActionCreate},
		{Object: source[types.ObjectKey(types.KindTable, "clientes")], Action: types.ActionCreate},
		{Object: source[types.ObjectKey(types.KindTable, "empresa")], Action: types.ActionCreate},
		{Object: target[types.ObjectKey(types.KindTable, "empresa")], Action: types.ActionUpdate},
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
	if !strings.Contains(sqlText, "CREATE TABLE clientes (id INT);") {
		t.Fatalf("esperava CREATE TABLE, recebi: %s", sqlText)
	}
	if strings.Contains(sqlText, "DROP TABLE IF EXISTS `empresa`;") {
		t.Fatalf("não esperava DROP TABLE, recebi: %s", sqlText)
	}
	if !strings.Contains(sqlText, "alter table empresa add column criado_em timestamp null;") {
		t.Fatalf("esperava ALTER TABLE, recebi: %s", sqlText)
	}
}
