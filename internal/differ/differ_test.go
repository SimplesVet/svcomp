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

func TestSplitAlterForFKConflict(t *testing.T) {
	// Simula o caso reportado: schemadiff gera um ALTER TABLE que faz
	// DROP FOREIGN KEY fk_x e ADD CONSTRAINT fk_x na mesma instrução,
	// causando o erro MySQL 1826.
	alterSQL := "alter table ajuda " +
		"drop foreign key fk_ajuda_documentacao, " +
		"drop key ak_ajuda_ajudachave, " +
		"modify column ajudachave varchar(40) not null, " +
		"add constraint fk_ajuda_documentacao foreign key (documentacao_id) references documentacao (documentacao_id)"

	result := splitAlterForFKConflict(alterSQL)

	parts := strings.Split(strings.TrimSpace(result), "\n")
	if len(parts) != 2 {
		t.Fatalf("esperava 2 instruções separadas, mas recebi %d: %s", len(parts), result)
	}

	stmt1 := strings.ToLower(parts[0])
	stmt2 := strings.ToLower(parts[1])

	if !strings.Contains(stmt1, "drop foreign key") {
		t.Errorf("1º statement deveria conter DROP FOREIGN KEY: %s", stmt1)
	}
	if !strings.Contains(stmt1, "drop key") {
		t.Errorf("1º statement deveria conter DROP KEY: %s", stmt1)
	}
	if strings.Contains(stmt1, "add constraint") {
		t.Errorf("1º statement não deveria conter ADD CONSTRAINT: %s", stmt1)
	}

	if strings.Contains(stmt2, "drop foreign key") {
		t.Errorf("2º statement não deveria conter DROP FOREIGN KEY: %s", stmt2)
	}
	if !strings.Contains(stmt2, "modify column") {
		t.Errorf("2º statement deveria conter MODIFY COLUMN: %s", stmt2)
	}
	if !strings.Contains(stmt2, "add constraint fk_ajuda_documentacao foreign key") {
		t.Errorf("2º statement deveria conter ADD CONSTRAINT com o mesmo nome: %s", stmt2)
	}
}

func TestSplitAlterForFKConflict_NoConflict(t *testing.T) {
	// Quando não há conflito de nomes, o ALTER TABLE não deve ser dividido.
	alterSQL := "alter table produtos " +
		"drop foreign key fk_categoria, " +
		"add constraint fk_fornecedor foreign key (fornecedor_id) references fornecedores (id)"

	result := splitAlterForFKConflict(alterSQL)

	parts := strings.Split(strings.TrimSpace(result), "\n")
	if len(parts) != 1 {
		t.Fatalf("sem conflito de nomes: esperava 1 instrução, recebi %d: %s", len(parts), result)
	}
}
