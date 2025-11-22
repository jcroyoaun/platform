-- Enable UUID extension for unique IDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ENUMS for strict type checking on calculations
CREATE TYPE periodicity_enum AS ENUM ('DAILY', 'WEEKLY', 'BIWEEKLY', 'SEMIMONTHLY', 'MONTHLY', 'ANNUAL');
CREATE TYPE infonavit_credit_type AS ENUM ('PERCENTAGE', 'FIXED_PESOS', 'VSM');

-- 1. FISCAL CONFIGURATION (The "Source of Truth")
-- One row per year. Holds all the global constants.
CREATE TABLE fiscal_years (
    id SERIAL PRIMARY KEY,
    year INT NOT NULL UNIQUE, -- e.g., 2025
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Unit References (Daily Values)
    uma_daily NUMERIC(10, 4) NOT NULL,   -- 2025: $113.14
    uma_monthly NUMERIC(10, 4) NOT NULL, -- 2025: $3,439.46
    uma_annual NUMERIC(12, 4) NOT NULL,  -- 2025: $41,273.52
    umi_value NUMERIC(10, 4) NOT NULL,   -- "Unidad Mixta Infonavit" for VSM credits
    
    -- Minimum Wages (Daily)
    smg_general NUMERIC(10, 2) NOT NULL, -- 2025: $278.80
    smg_border NUMERIC(10, 2) NOT NULL,  -- 2025: $419.88
    
    -- 2025 Subsidio al Empleo Logic (It's now a factor, not a table)
    subsidy_factor NUMERIC(6, 4) DEFAULT 0.1380,    -- 13.8%
    subsidy_threshold_monthly NUMERIC(10, 2) DEFAULT 10171.00, -- The "Cliff" limit
    
    -- Benefit Legal Limits (Capped Deductions)
    -- Fondo de Ahorro: Capped at 1.3 times the UMA
    fa_legal_cap_uma_factor NUMERIC(4, 2) DEFAULT 1.30, 
    -- Fondo de Ahorro: Capped at 13% of salary
    fa_legal_max_percentage NUMERIC(5, 4) DEFAULT 0.1300, 
    
    -- Vales de Despensa: Capped at 40% UMA for IMSS Integration exclusion
    pantry_vouchers_uma_cap NUMERIC(4, 2) DEFAULT 0.40,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 2. ISR TAX BRACKETS (Tarifas ISR)
-- Stores the progressive tax table.
CREATE TABLE isr_brackets (
    id SERIAL PRIMARY KEY,
    fiscal_year_id INT REFERENCES fiscal_years(id),
    periodicity periodicity_enum DEFAULT 'MONTHLY',
    
    lower_limit NUMERIC(12, 2) NOT NULL,
    upper_limit NUMERIC(12, 2) NOT NULL, -- Use 999999999 for "En Adelante"
    fixed_fee NUMERIC(12, 2) NOT NULL,   -- Cuota Fija
    surplus_percent NUMERIC(6, 4) NOT NULL -- Excedente (store 0.1088 for 10.88%)
);

-- 3. SENIORITY & INTEGRATION (Vacaciones Dignas)
-- Used to calculate the "Factor de Integración" for SBC
CREATE TABLE seniority_benefits (
    id SERIAL PRIMARY KEY,
    fiscal_year_id INT REFERENCES fiscal_years(id),
    
    years_of_service INT NOT NULL, -- 1, 2, 3...
    vacation_days INT NOT NULL,    -- 12, 14, 16...
    prima_vacacional_percent NUMERIC(5, 4) DEFAULT 0.25, -- Minimum 25%
    aguinaldo_days INT DEFAULT 15, -- Minimum 15 days
    
    -- Constraint to ensure we don't have duplicate years for the same fiscal config
    UNIQUE(fiscal_year_id, years_of_service)
);

-- 4. IMSS CONFIGURATION (Fixed Percentages)
-- Standard Worker/Employer quotas (Enfermedad, Invalidez, Guarderias, Retiro)
CREATE TABLE imss_concepts (
    id SERIAL PRIMARY KEY,
    concept_name VARCHAR(100) NOT NULL, -- e.g. "Enfermedad y Maternidad (Gastos Médicos)"
    
    worker_percent NUMERIC(7, 5) DEFAULT 0,
    employer_percent NUMERIC(7, 5) DEFAULT 0,
    
    -- Some concepts are capped at 25 UMA, others (Invalidez) might differ in base
    base_cap_in_umas INT DEFAULT 25, 
    
    is_fixed_rate BOOLEAN DEFAULT TRUE 
    -- If FALSE, it means we must look up a progressive table (like Cesantía 2025)
);

-- 5. IMSS EMPLOYER PROGRESSIVE TABLE (Cesantía en Edad Avanzada)
-- The new 2025 table where Employer pays more based on how high the salary is
CREATE TABLE imss_employer_cesantia_brackets (
    id SERIAL PRIMARY KEY,
    fiscal_year_id INT REFERENCES fiscal_years(id),
    
    lower_bound_uma NUMERIC(6, 3) NOT NULL, -- e.g., 1.01 UMA
    upper_bound_uma NUMERIC(6, 3) NOT NULL, -- e.g., 1.50 UMA
    employer_percent NUMERIC(7, 5) NOT NULL -- The rate (e.g., 0.03544)
);

-- 6. USER INPUTS / SIMULATIONS
-- Where the user data lives
CREATE TABLE salary_inputs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fiscal_year_id INT REFERENCES fiscal_years(id),
    created_at TIMESTAMP DEFAULT NOW(),
    
    -- Basic Info
    gross_monthly_salary NUMERIC(12, 2) NOT NULL,
    zip_code VARCHAR(10), -- To detect Border Zone (ZLFN)
    years_of_service INT DEFAULT 0, -- To lookup vacation factor
    
    -- Company "Superior" Benefits (User overrides)
    company_aguinaldo_days INT DEFAULT 15,
    company_vacation_days INT DEFAULT 12,
    company_prima_vacacional_percent NUMERIC(5, 4) DEFAULT 0.25,
    
    -- Fondo de Ahorro Logic
    has_fondo_ahorro BOOLEAN DEFAULT FALSE,
    fondo_ahorro_company_percent NUMERIC(5, 4) DEFAULT 0, -- e.g. 0.13
    fondo_ahorro_employee_percent NUMERIC(5, 4) DEFAULT 0,
    
    -- Vales
    has_pantry_vouchers BOOLEAN DEFAULT FALSE,
    pantry_vouchers_amount NUMERIC(10, 2) DEFAULT 0,
    
    -- Infonavit
    has_infonavit_loan BOOLEAN DEFAULT FALSE,
    infonavit_type infonavit_credit_type,
    infonavit_value NUMERIC(10, 4) -- Can be %, Fixed Pesos, or Factor VSM
);

