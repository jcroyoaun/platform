-- =====================================================
-- Migration Rollback: Remove RESICO 2025 Data
-- =====================================================

-- Delete RESICO retention rules for 2025
DELETE FROM fiscal_retention_rules 
WHERE fiscal_year_id = (SELECT id FROM fiscal_years WHERE year = 2025 LIMIT 1)
  AND regime_name = 'RESICO';

-- Delete RESICO tax brackets for 2025
DELETE FROM resico_brackets 
WHERE fiscal_year_id = (SELECT id FROM fiscal_years WHERE year = 2025 LIMIT 1);

-- Verify deletion
DO $$
DECLARE
    bracket_count INTEGER;
    retention_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO bracket_count FROM resico_brackets WHERE fiscal_year_id = (SELECT id FROM fiscal_years WHERE year = 2025 LIMIT 1);
    SELECT COUNT(*) INTO retention_count FROM fiscal_retention_rules WHERE fiscal_year_id = (SELECT id FROM fiscal_years WHERE year = 2025 LIMIT 1) AND regime_name = 'RESICO';
    
    RAISE NOTICE 'RESICO 2025 Data Removed:';
    RAISE NOTICE '  - Tax Brackets remaining: %', bracket_count;
    RAISE NOTICE '  - Retention Rules remaining: %', retention_count;
END $$;

