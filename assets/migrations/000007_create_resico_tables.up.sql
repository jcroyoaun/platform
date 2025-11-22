-- =====================================================
-- Migration: Create RESICO Tables
-- Description: Tables for RESICO tax regime calculations
-- =====================================================

-- RESICO Tax Brackets (Flat Rate per Tier)
-- Unlike ISR, RESICO applies a flat rate to the TOTAL income (not surplus)
CREATE TABLE resico_brackets (
    id SERIAL PRIMARY KEY,
    fiscal_year_id INT REFERENCES fiscal_years(id) ON DELETE CASCADE,
    periodicity periodicity_enum DEFAULT 'MONTHLY',
    
    -- The tier ceiling
    upper_limit NUMERIC(12, 2) NOT NULL, 
    
    -- The fixed rate for this entire tier (e.g., 0.0100 for 1%)
    -- NOTE: In RESICO, you pay this rate on the TOTAL income
    applicable_rate NUMERIC(5, 4) NOT NULL,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create index for faster lookups
CREATE INDEX idx_resico_brackets_fiscal_year ON resico_brackets(fiscal_year_id);

-- RESICO Retention Rules
-- When you bill a client, they may retain ISR depending on their type
CREATE TABLE fiscal_retention_rules (
    id SERIAL PRIMARY KEY,
    fiscal_year_id INT REFERENCES fiscal_years(id) ON DELETE CASCADE,
    
    regime_name VARCHAR(50) DEFAULT 'RESICO',
    client_type VARCHAR(50) CHECK (client_type IN ('PERSONA_FISICA', 'PERSONA_MORAL')),
    
    -- The retention rate (e.g., 0.0125 for 1.25%)
    retention_rate NUMERIC(5, 4) DEFAULT 0.0, 
    
    description VARCHAR(200),
    created_at TIMESTAMP DEFAULT NOW(),
    
    -- Ensure unique combination
    UNIQUE(fiscal_year_id, regime_name, client_type)
);

-- Create index for faster lookups
CREATE INDEX idx_retention_rules_fiscal_year ON fiscal_retention_rules(fiscal_year_id);

-- Comments for documentation
COMMENT ON TABLE resico_brackets IS 'RESICO tax brackets with flat rates applied to total income';
COMMENT ON COLUMN resico_brackets.applicable_rate IS 'Flat rate applied to TOTAL income (not surplus like ISR)';
COMMENT ON TABLE fiscal_retention_rules IS 'Retention rules for different client types under RESICO';
COMMENT ON COLUMN fiscal_retention_rules.retention_rate IS 'Percentage retained by client when issuing payment';

