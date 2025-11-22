-- =====================================================
-- Migration: Seed RESICO 2025 Data
-- Description: Tax brackets and retention rules for RESICO 2025
-- =====================================================

-- Insert RESICO Monthly Tax Brackets for 2025
-- Note: RESICO applies a FLAT RATE to the TOTAL income (not progressive like ISR)
INSERT INTO resico_brackets (fiscal_year_id, periodicity, upper_limit, applicable_rate) VALUES
    -- Get fiscal_year_id for 2025
    ((SELECT id FROM fiscal_years WHERE year = 2025 LIMIT 1), 'MONTHLY', 25000.00, 0.0100),   -- Up to $25k -> 1.00%
    ((SELECT id FROM fiscal_years WHERE year = 2025 LIMIT 1), 'MONTHLY', 50000.00, 0.0110),   -- Up to $50k -> 1.10%
    ((SELECT id FROM fiscal_years WHERE year = 2025 LIMIT 1), 'MONTHLY', 83333.33, 0.0150),   -- Up to $83.3k -> 1.50%
    ((SELECT id FROM fiscal_years WHERE year = 2025 LIMIT 1), 'MONTHLY', 208333.33, 0.0200),  -- Up to $208.3k -> 2.00%
    ((SELECT id FROM fiscal_years WHERE year = 2025 LIMIT 1), 'MONTHLY', 9999999.99, 0.0250); -- Above $208.3k -> 2.50% (annual cap is $3.5M)

-- Insert RESICO Retention Rules for 2025
-- When you invoice a client, they retain ISR based on their entity type
INSERT INTO fiscal_retention_rules (fiscal_year_id, regime_name, client_type, retention_rate, description) VALUES
    (
        (SELECT id FROM fiscal_years WHERE year = 2025 LIMIT 1),
        'RESICO',
        'PERSONA_MORAL',
        0.0125,  -- 1.25% retention
        'Retención del 1.25% cuando el cliente es Persona Moral (empresa)'
    ),
    (
        (SELECT id FROM fiscal_years WHERE year = 2025 LIMIT 1),
        'RESICO',
        'PERSONA_FISICA',
        0.0000,  -- 0.00% retention (no retention)
        'Sin retención cuando el cliente es Persona Física (freelancer a freelancer)'
    );

-- Verify the data was inserted
DO $$
DECLARE
    bracket_count INTEGER;
    retention_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO bracket_count FROM resico_brackets WHERE fiscal_year_id = (SELECT id FROM fiscal_years WHERE year = 2025 LIMIT 1);
    SELECT COUNT(*) INTO retention_count FROM fiscal_retention_rules WHERE fiscal_year_id = (SELECT id FROM fiscal_years WHERE year = 2025 LIMIT 1);
    
    RAISE NOTICE 'RESICO 2025 Data Seeded:';
    RAISE NOTICE '  - Tax Brackets: % rows', bracket_count;
    RAISE NOTICE '  - Retention Rules: % rows', retention_count;
    
    IF bracket_count != 5 THEN
        RAISE EXCEPTION 'Expected 5 RESICO brackets, but found %', bracket_count;
    END IF;
    
    IF retention_count != 2 THEN
        RAISE EXCEPTION 'Expected 2 retention rules, but found %', retention_count;
    END IF;
END $$;

