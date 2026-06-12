//go:build integration

package lister

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/simplesvet/svcomp/internal/connector"
	"github.com/simplesvet/svcomp/pkg/types"
)

func TestIntegrationListAllSourceSchema(t *testing.T) {
	objs := listFromDSN(t, envOrDefault("INTEGRATION_SOURCE_DSN", "root:root@tcp(127.0.0.1:3307)/svcomp_source?parseTime=true"))

	expected := map[string]struct{}{
		types.ObjectKey(types.KindTable, "users"):              {},
		types.ObjectKey(types.KindTable, "pets"):               {},
		types.ObjectKey(types.KindTable, "audit_logs"):         {},
		types.ObjectKey(types.KindView, "v_active_users"):      {},
		types.ObjectKey(types.KindFunction, "fn_total_pets"):   {},
		types.ObjectKey(types.KindProcedure, "sp_touch_users"): {},
		types.ObjectKey(types.KindTrigger, "trg_pets_bi"):      {},
		types.ObjectKey(types.KindEvent, "ev_cleanup_audit"):   {},
	}

	assertObjectsPresent(t, objs, expected)
}

func TestIntegrationListAllTargetSchema(t *testing.T) {
	objs := listFromDSN(t, envOrDefault("INTEGRATION_TARGET_DSN", "root:root@tcp(127.0.0.1:3308)/svcomp_target?parseTime=true"))

	expected := map[string]struct{}{
		types.ObjectKey(types.KindTable, "users"):         {},
		types.ObjectKey(types.KindView, "v_active_users"): {},
	}

	assertObjectsPresent(t, objs, expected)
}

func listFromDSN(t *testing.T, dsn string) []types.DBObject {
	t.Helper()

	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("dsn inválida: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := connector.New(ctx, dsn)
	if err != nil {
		t.Fatalf("erro ao conectar: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("erro ao fechar conexão: %v", err)
		}
	}()

	objs, err := New(conn.DB(), cfg.DBName).ListAll(ctx)
	if err != nil {
		t.Fatalf("erro ao listar objetos: %v", err)
	}

	return objs
}

func assertObjectsPresent(t *testing.T, objs []types.DBObject, expected map[string]struct{}) {
	t.Helper()

	seen := make(map[string]struct{}, len(objs))
	for _, obj := range objs {
		seen[obj.Key()] = struct{}{}
	}

	for key := range expected {
		if _, ok := seen[key]; !ok {
			t.Fatalf("objeto esperado não encontrado: %s", key)
		}
	}
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
