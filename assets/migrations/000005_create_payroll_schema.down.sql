-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS salary_inputs;
DROP TABLE IF EXISTS imss_employer_cesantia_brackets;
DROP TABLE IF EXISTS imss_concepts;
DROP TABLE IF EXISTS seniority_benefits;
DROP TABLE IF EXISTS isr_brackets;
DROP TABLE IF EXISTS fiscal_years;

-- Drop custom types
DROP TYPE IF EXISTS infonavit_credit_type;
DROP TYPE IF EXISTS periodicity_enum;

-- Note: We're not dropping the uuid-ossp extension as it might be used by other parts of the application

