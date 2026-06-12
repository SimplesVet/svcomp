package lister

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/simplesvet/svcomp/pkg/types"
)

type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type Lister struct {
	db     Queryer
	dbName string
}

func New(db Queryer, dbName string) *Lister {
	return &Lister{db: db, dbName: dbName}
}

func (l *Lister) ListAll(ctx context.Context) ([]types.DBObject, error) {
	all := make([]types.DBObject, 0)

	tables, err := l.listTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("lister: tables: %w", err)
	}
	all = append(all, tables...)

	views, err := l.listViews(ctx)
	if err != nil {
		return nil, fmt.Errorf("lister: views: %w", err)
	}
	all = append(all, views...)

	functions, err := l.listFunctions(ctx)
	if err != nil {
		return nil, fmt.Errorf("lister: functions: %w", err)
	}
	all = append(all, functions...)

	procedures, err := l.listProcedures(ctx)
	if err != nil {
		return nil, fmt.Errorf("lister: procedures: %w", err)
	}
	all = append(all, procedures...)

	triggers, err := l.listTriggers(ctx)
	if err != nil {
		return nil, fmt.Errorf("lister: triggers: %w", err)
	}
	all = append(all, triggers...)

	events, err := l.listEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("lister: events: %w", err)
	}
	all = append(all, events...)

	return all, nil
}

func (l *Lister) listTables(ctx context.Context) ([]types.DBObject, error) {
	names, err := l.listNamesFromQuery(ctx, "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'", l.dbName)
	if err != nil {
		return nil, err
	}
	objects := make([]types.DBObject, 0, len(names))
	for _, name := range names {
		definition, err := l.showCreateOne(ctx, fmt.Sprintf("SHOW CREATE TABLE %s", quoteIdent(name)), "Create Table")
		if err != nil {
			return nil, fmt.Errorf("show create table %s: %w", name, err)
		}
		objects = append(objects, types.DBObject{Name: name, Kind: types.KindTable, Definition: definition})
	}
	return objects, nil
}

func (l *Lister) listViews(ctx context.Context) ([]types.DBObject, error) {
	names, err := l.listNamesFromQuery(ctx, "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.VIEWS WHERE TABLE_SCHEMA = ?", l.dbName)
	if err != nil {
		return nil, err
	}
	objects := make([]types.DBObject, 0, len(names))
	for _, name := range names {
		definition, err := l.showCreateOne(ctx, fmt.Sprintf("SHOW CREATE VIEW %s", quoteIdent(name)), "Create View")
		if err != nil {
			return nil, fmt.Errorf("show create view %s: %w", name, err)
		}
		objects = append(objects, types.DBObject{Name: name, Kind: types.KindView, Definition: definition})
	}
	return objects, nil
}

func (l *Lister) listFunctions(ctx context.Context) ([]types.DBObject, error) {
	names, err := l.listNamesFromQuery(ctx, "SELECT ROUTINE_NAME FROM INFORMATION_SCHEMA.ROUTINES WHERE ROUTINE_SCHEMA = ? AND ROUTINE_TYPE = 'FUNCTION'", l.dbName)
	if err != nil {
		return nil, err
	}
	objects := make([]types.DBObject, 0, len(names))
	for _, name := range names {
		definition, err := l.showCreateOne(ctx, fmt.Sprintf("SHOW CREATE FUNCTION %s", quoteIdent(name)), "Create Function")
		if err != nil {
			return nil, fmt.Errorf("show create function %s: %w", name, err)
		}
		objects = append(objects, types.DBObject{Name: name, Kind: types.KindFunction, Definition: definition})
	}
	return objects, nil
}

func (l *Lister) listProcedures(ctx context.Context) ([]types.DBObject, error) {
	names, err := l.listNamesFromQuery(ctx, "SELECT ROUTINE_NAME FROM INFORMATION_SCHEMA.ROUTINES WHERE ROUTINE_SCHEMA = ? AND ROUTINE_TYPE = 'PROCEDURE'", l.dbName)
	if err != nil {
		return nil, err
	}
	objects := make([]types.DBObject, 0, len(names))
	for _, name := range names {
		definition, err := l.showCreateOne(ctx, fmt.Sprintf("SHOW CREATE PROCEDURE %s", quoteIdent(name)), "Create Procedure")
		if err != nil {
			return nil, fmt.Errorf("show create procedure %s: %w", name, err)
		}
		objects = append(objects, types.DBObject{Name: name, Kind: types.KindProcedure, Definition: definition})
	}
	return objects, nil
}

func (l *Lister) listTriggers(ctx context.Context) ([]types.DBObject, error) {
	names, err := l.listNamesFromQuery(ctx, "SELECT TRIGGER_NAME FROM INFORMATION_SCHEMA.TRIGGERS WHERE TRIGGER_SCHEMA = ?", l.dbName)
	if err != nil {
		return nil, err
	}
	objects := make([]types.DBObject, 0, len(names))
	for _, name := range names {
		definition, err := l.showCreateOne(ctx, fmt.Sprintf("SHOW CREATE TRIGGER %s", quoteIdent(name)), "SQL Original Statement")
		if err != nil {
			return nil, fmt.Errorf("show create trigger %s: %w", name, err)
		}
		objects = append(objects, types.DBObject{Name: name, Kind: types.KindTrigger, Definition: definition})
	}
	return objects, nil
}

func (l *Lister) listEvents(ctx context.Context) ([]types.DBObject, error) {
	names, err := l.listNamesFromQuery(ctx, "SELECT EVENT_NAME FROM INFORMATION_SCHEMA.EVENTS WHERE EVENT_SCHEMA = ?", l.dbName)
	if err != nil {
		return nil, err
	}
	objects := make([]types.DBObject, 0, len(names))
	for _, name := range names {
		definition, err := l.showCreateOne(ctx, fmt.Sprintf("SHOW CREATE EVENT %s", quoteIdent(name)), "Create Event")
		if err != nil {
			return nil, fmt.Errorf("show create event %s: %w", name, err)
		}
		objects = append(objects, types.DBObject{Name: name, Kind: types.KindEvent, Definition: definition})
	}
	return objects, nil
}

func (l *Lister) listNamesFromQuery(ctx context.Context, query string, args ...any) ([]string, error) {
	rows, err := l.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return names, nil
}

func (l *Lister) showCreateOne(ctx context.Context, query string, createColumn string) (string, error) {
	rows, err := l.db.QueryContext(ctx, query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}
	values := make([]any, len(columns))
	raw := make([]sql.RawBytes, len(columns))
	for i := range raw {
		values[i] = &raw[i]
	}

	if !rows.Next() {
		return "", sql.ErrNoRows
	}
	if err := rows.Scan(values...); err != nil {
		return "", err
	}

	idx := -1
	for i, col := range columns {
		if strings.EqualFold(col, createColumn) {
			idx = i
			break
		}
	}
	if idx == -1 {
		return "", fmt.Errorf("coluna %q não encontrada em SHOW CREATE", createColumn)
	}
	return string(raw[idx]), nil
}

func quoteIdent(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}
