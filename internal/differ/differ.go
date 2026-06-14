package differ

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"vitess.io/vitess/go/vt/schemadiff"
	"vitess.io/vitess/go/vt/sqlparser"

	"github.com/simplesvet/svcomp/pkg/types"
)

type Generator struct {
}

func NewGenerator() (*Generator, error) {
	return &Generator{}, nil
}

func (g *Generator) Generate(ctx context.Context, diffs []types.DiffResult, sourceIndex, targetIndex map[string]types.DBObject) (string, error) {
	_ = ctx
	ordered := make([]types.DiffResult, len(diffs))
	copy(ordered, diffs)
	sort.SliceStable(ordered, func(i, j int) bool {
		io := kindOrder(ordered[i].Object.Kind)
		jo := kindOrder(ordered[j].Object.Kind)
		if io != jo {
			return io < jo
		}
		if ordered[i].Object.Name != ordered[j].Object.Name {
			return ordered[i].Object.Name < ordered[j].Object.Name
		}
		return actionOrder(ordered[i].Action) < actionOrder(ordered[j].Action)
	})

	var out strings.Builder
	for _, diff := range ordered {
		sqlBlock, err := g.sqlForDiff(ctx, diff, sourceIndex, targetIndex)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(sqlBlock) == "" {
			continue
		}
		out.WriteString(commentFor(diff))
		out.WriteString("\n")
		out.WriteString(sqlBlock)
		if !strings.HasSuffix(sqlBlock, "\n") {
			out.WriteString("\n")
		}
		out.WriteString("\n")
	}

	return strings.TrimSpace(out.String()) + "\n", nil
}

func (g *Generator) sqlForDiff(ctx context.Context, diff types.DiffResult, sourceIndex, targetIndex map[string]types.DBObject) (string, error) {
	switch diff.Object.Kind {
	case types.KindTable:
		return g.sqlForTableDiff(ctx, diff, sourceIndex, targetIndex)
	default:
		return sqlForNonTable(diff), nil
	}
}

func (g *Generator) sqlForTableDiff(ctx context.Context, diff types.DiffResult, sourceIndex, targetIndex map[string]types.DBObject) (string, error) {
	name := diff.Object.Name
	key := diff.Object.Key()
	switch diff.Action {
	case types.ActionDelete:
		return fmt.Sprintf("DROP TABLE IF EXISTS %s;", quoteIdent(name)), nil
	case types.ActionCreate:
		src := sourceIndex[key]
		return ensureSemicolon(src.Definition), nil
	case types.ActionUpdate:
		src, okSrc := sourceIndex[key]
		tgt, okTgt := targetIndex[key]
		if !okSrc || !okTgt {
			return "", fmt.Errorf("differ: tabela %s sem definição source/target para update", name)
		}

		alterSQL, err := diffTableUsingPackage(tgt.Definition, src.Definition)
		if err == nil && strings.TrimSpace(alterSQL) != "" {
			return splitAlterForFKConflict(alterSQL), nil
		}

		if normalizeDDL(src.Definition) == normalizeDDL(tgt.Definition) {
			return "", nil
		}

		// Fallback seguro quando o schemadiff não consegue gerar o ALTER.
		if err != nil {
			fmt.Printf("differ: aviso: não foi possível gerar ALTER TABLE para %s usando schemadiff. Um DROP foi gerado: %v\n", name, err)
		}
		return fmt.Sprintf("DROP TABLE IF EXISTS %s;\n%s", quoteIdent(name), ensureSemicolon(src.Definition)), nil
	default:
		return "", fmt.Errorf("differ: ação inválida para tabela %s: %s", name, diff.Action)
	}
}

// splitAlterForFKConflict resolve o erro MySQL 1826 (Duplicate foreign key
// constraint name) que ocorre quando um único ALTER TABLE faz DROP FOREIGN KEY
// e ADD CONSTRAINT com o mesmo nome na mesma instrução. Quando detecta esse
// conflito, divide em dois ALTER TABLE: primeiro os drops (DropKey), depois o
// restante (MODIFY, ADD, etc.).
func splitAlterForFKConflict(alterSQL string) string {
	p := sqlparser.NewTestParser()
	stmt, err := p.Parse(alterSQL)
	if err != nil {
		return ensureSemicolon(alterSQL)
	}

	alterStmt, ok := stmt.(*sqlparser.AlterTable)
	if !ok {
		return ensureSemicolon(alterSQL)
	}

	// Coleta nomes de FKs sendo removidas.
	droppedFKNames := make(map[string]bool)
	for _, opt := range alterStmt.AlterOptions {
		if dk, ok := opt.(*sqlparser.DropKey); ok && dk.Type == sqlparser.ForeignKeyType {
			droppedFKNames[dk.Name.Lowered()] = true
		}
	}

	if len(droppedFKNames) == 0 {
		return ensureSemicolon(alterSQL)
	}

	// Verifica se alguma FK sendo adicionada tem o mesmo nome de uma sendo removida.
	hasConflict := false
	for _, opt := range alterStmt.AlterOptions {
		if acd, ok := opt.(*sqlparser.AddConstraintDefinition); ok {
			if _, isFKDef := acd.ConstraintDefinition.Details.(*sqlparser.ForeignKeyDefinition); isFKDef {
				if droppedFKNames[acd.ConstraintDefinition.Name.Lowered()] {
					hasConflict = true
					break
				}
			}
		}
	}

	if !hasConflict {
		return ensureSemicolon(alterSQL)
	}

	// Divide: todos os DropKey vão para o primeiro statement; o restante para o segundo.
	var dropOpts, restOpts []sqlparser.AlterOption
	for _, opt := range alterStmt.AlterOptions {
		if _, ok := opt.(*sqlparser.DropKey); ok {
			dropOpts = append(dropOpts, opt)
		} else {
			restOpts = append(restOpts, opt)
		}
	}

	if len(dropOpts) == 0 || len(restOpts) == 0 {
		return ensureSemicolon(alterSQL)
	}

	stmt1 := *alterStmt
	stmt1.AlterOptions = dropOpts
	stmt2 := *alterStmt
	stmt2.AlterOptions = restOpts

	return ensureSemicolon(sqlparser.String(&stmt1)) + "\n" + ensureSemicolon(sqlparser.String(&stmt2))
}

func diffTableUsingPackage(tgtDDL, srcDDL string) (string, error) {
	env := schemadiff.NewTestEnv()
	diff, err := schemadiff.DiffCreateTablesQueries(env, tgtDDL, srcDDL, schemadiff.EmptyDiffHints())
	if err != nil {
		return "", fmt.Errorf("differ: schemadiff: %w", err)
	}
	if diff == nil || diff.IsEmpty() {
		return "", nil
	}
	return diff.StatementString(), nil
}

func sqlForNonTable(diff types.DiffResult) string {
	name := quoteIdent(diff.Object.Name)
	kind := ddlKind(diff.Object.Kind)
	createStmt := ensureSemicolon(diff.Object.Definition)
	dropStmt := fmt.Sprintf("DROP %s IF EXISTS %s;", kind, name)

	switch diff.Action {
	case types.ActionDelete:
		return dropStmt
	case types.ActionCreate, types.ActionUpdate:
		return dropStmt + "\n" + createStmt
	default:
		return ""
	}
}

func commentFor(diff types.DiffResult) string {
	return fmt.Sprintf("-- [%s] %s: %s", strings.ToUpper(string(diff.Object.Kind)), diff.Object.Name, strings.ToUpper(string(diff.Action)))
}

func ddlKind(kind types.ObjectKind) string {
	switch kind {
	case types.KindView:
		return "VIEW"
	case types.KindFunction:
		return "FUNCTION"
	case types.KindProcedure:
		return "PROCEDURE"
	case types.KindTrigger:
		return "TRIGGER"
	case types.KindEvent:
		return "EVENT"
	default:
		return strings.ToUpper(string(kind))
	}
}

func ensureSemicolon(sql string) string {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return ""
	}
	if strings.HasSuffix(trimmed, ";") {
		return trimmed
	}
	return trimmed + ";"
}

func normalizeDDL(sql string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(sql), " "))
}

func quoteIdent(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func kindOrder(kind types.ObjectKind) int {
	for i, k := range types.OrderedKinds {
		if k == kind {
			return i
		}
	}
	return len(types.OrderedKinds) + 1
}

func actionOrder(action types.DiffAction) int {
	switch action {
	case types.ActionDelete:
		return 0
	case types.ActionUpdate:
		return 1
	case types.ActionCreate:
		return 2
	default:
		return 3
	}
}
