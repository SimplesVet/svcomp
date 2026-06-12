package main

import (
	"context"
	"fmt"
	"time"

	"github.com/simplesvet/svcomp/internal/comparator"
	"github.com/simplesvet/svcomp/internal/config"
	"github.com/simplesvet/svcomp/internal/connector"
	"github.com/simplesvet/svcomp/internal/differ"
	"github.com/simplesvet/svcomp/internal/lister"
	"github.com/simplesvet/svcomp/internal/output"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "svcomp",
		Short: "Compara schemas MySQL e gera SQL de sincronização",
	}
	cmd.AddCommand(newCompareCmd())
	return cmd
}

func newCompareCmd() *cobra.Command {
	cfg := &config.Config{}
	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Compara source e target e gera SQL",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompare(cmd.Context(), cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.SourceDSN, "source", "", "DSN do banco source")
	cmd.Flags().StringVar(&cfg.TargetDSN, "target", "", "DSN do banco target")
	cmd.Flags().StringVar(&cfg.OutputFile, "output", "", "Arquivo de saída SQL")
	cmd.Flags().BoolVar(&cfg.DryRun, "dry-run", false, "Apenas imprime o SQL gerado")
	_ = cmd.MarkFlagRequired("source")
	_ = cmd.MarkFlagRequired("target")

	return cmd
}

func runCompare(ctx context.Context, cfg *config.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	sourceConn, err := connector.New(ctx, cfg.SourceDSN)
	if err != nil {
		return fmt.Errorf("compare: conexão source: %w", err)
	}
	defer sourceConn.Close()

	targetConn, err := connector.New(ctx, cfg.TargetDSN)
	if err != nil {
		return fmt.Errorf("compare: conexão target: %w", err)
	}
	defer targetConn.Close()

	sourceObjects, err := lister.New(sourceConn.DB(), cfg.SourceDBName).ListAll(ctx)
	if err != nil {
		return fmt.Errorf("compare: listar source: %w", err)
	}
	targetObjects, err := lister.New(targetConn.DB(), cfg.TargetDBName).ListAll(ctx)
	if err != nil {
		return fmt.Errorf("compare: listar target: %w", err)
	}

	cmp := comparator.New()
	diffs := cmp.Compare(sourceObjects, targetObjects)

	generator, err := differ.NewGenerator()
	if err != nil {
		return err
	}

	sqlText, err := generator.Generate(ctx, diffs, comparator.IndexByKey(sourceObjects), comparator.IndexByKey(targetObjects))
	if err != nil {
		return fmt.Errorf("compare: gerar SQL: %w", err)
	}

	if cfg.DryRun {
		return output.WriteSQL(sqlText, "")
	}
	return output.WriteSQL(sqlText, cfg.OutputFile)
}
