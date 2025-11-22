-- =====================================================
-- Migration Rollback: Drop RESICO Tables
-- =====================================================

DROP INDEX IF EXISTS idx_retention_rules_fiscal_year;
DROP INDEX IF EXISTS idx_resico_brackets_fiscal_year;

DROP TABLE IF EXISTS fiscal_retention_rules;
DROP TABLE IF EXISTS resico_brackets;

