-- Remove 2025 seeded data
DELETE FROM imss_employer_cesantia_brackets WHERE fiscal_year_id = (SELECT id FROM fiscal_years WHERE year = 2025);
DELETE FROM imss_concepts;
DELETE FROM seniority_benefits WHERE fiscal_year_id = (SELECT id FROM fiscal_years WHERE year = 2025);
DELETE FROM isr_brackets WHERE fiscal_year_id = (SELECT id FROM fiscal_years WHERE year = 2025);
DELETE FROM fiscal_years WHERE year = 2025;

