package config

import (
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

type Config struct {
	SourceDSN    string
	TargetDSN    string
	OutputFile   string
	DryRun       bool
	SourceDBName string
	TargetDBName string
}

func (c *Config) Validate() error {
	if c.SourceDSN == "" {
		return errors.New("source DSN é obrigatório")
	}
	if c.TargetDSN == "" {
		return errors.New("target DSN é obrigatório")
	}

	sourceCfg, err := mysql.ParseDSN(c.SourceDSN)
	if err != nil {
		return fmt.Errorf("config: DSN source inválida: %w", err)
	}
	targetCfg, err := mysql.ParseDSN(c.TargetDSN)
	if err != nil {
		return fmt.Errorf("config: DSN target inválida: %w", err)
	}
	if sourceCfg.DBName == "" {
		return errors.New("config: DSN source deve conter nome do banco")
	}
	if targetCfg.DBName == "" {
		return errors.New("config: DSN target deve conter nome do banco")
	}

	c.SourceDBName = sourceCfg.DBName
	c.TargetDBName = targetCfg.DBName
	return nil
}
