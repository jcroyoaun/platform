-- Seed 2025 Fiscal Year Configuration
INSERT INTO fiscal_years (
    year, is_active,
    uma_daily, uma_monthly, uma_annual, umi_value,
    smg_general, smg_border,
    subsidy_factor, subsidy_threshold_monthly,
    fa_legal_cap_uma_factor, fa_legal_max_percentage,
    pantry_vouchers_uma_cap
) VALUES (
    2025, TRUE,
    113.14, 3439.46, 41273.52, 113.14,
    278.80, 419.88,
    0.1380, 10171.00,
    1.30, 0.1300,
    0.40
);

-- Seed ISR Brackets for 2025 (Monthly)
INSERT INTO isr_brackets (fiscal_year_id, periodicity, lower_limit, upper_limit, fixed_fee, surplus_percent) VALUES
((SELECT id FROM fiscal_years WHERE year = 2025), 'MONTHLY', 0.01, 746.04, 0.00, 0.0192),
((SELECT id FROM fiscal_years WHERE year = 2025), 'MONTHLY', 746.05, 6332.05, 14.32, 0.0640),
((SELECT id FROM fiscal_years WHERE year = 2025), 'MONTHLY', 6332.06, 11128.01, 371.83, 0.1088),
((SELECT id FROM fiscal_years WHERE year = 2025), 'MONTHLY', 11128.02, 12935.82, 893.63, 0.1600),
((SELECT id FROM fiscal_years WHERE year = 2025), 'MONTHLY', 12935.83, 15487.71, 1182.88, 0.1792),
((SELECT id FROM fiscal_years WHERE year = 2025), 'MONTHLY', 15487.72, 31236.49, 1640.18, 0.2136),
((SELECT id FROM fiscal_years WHERE year = 2025), 'MONTHLY', 31236.50, 49233.00, 5004.12, 0.2352),
((SELECT id FROM fiscal_years WHERE year = 2025), 'MONTHLY', 49233.01, 93993.90, 9236.89, 0.3000),
((SELECT id FROM fiscal_years WHERE year = 2025), 'MONTHLY', 93993.91, 125325.20, 22665.17, 0.3200),
((SELECT id FROM fiscal_years WHERE year = 2025), 'MONTHLY', 125325.21, 375975.61, 32691.18, 0.3400),
((SELECT id FROM fiscal_years WHERE year = 2025), 'MONTHLY', 375975.62, 999999999.99, 117912.32, 0.3500);

-- Seed Seniority Benefits (Vacaciones Dignas - Reforma 2023)
-- Years 1-25 for 2025
INSERT INTO seniority_benefits (fiscal_year_id, years_of_service, vacation_days, prima_vacacional_percent, aguinaldo_days) VALUES
((SELECT id FROM fiscal_years WHERE year = 2025), 1, 12, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 2, 14, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 3, 16, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 4, 18, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 5, 20, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 6, 22, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 7, 24, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 8, 26, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 9, 28, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 10, 30, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 11, 32, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 12, 34, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 13, 36, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 14, 38, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 15, 40, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 16, 42, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 17, 44, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 18, 46, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 19, 48, 0.25, 15),
((SELECT id FROM fiscal_years WHERE year = 2025), 20, 50, 0.25, 15);

-- Seed IMSS Concepts (Fixed Rates 2025)
INSERT INTO imss_concepts (concept_name, worker_percent, employer_percent, base_cap_in_umas, is_fixed_rate) VALUES
('Enfermedad y Maternidad (Gastos Médicos)', 0.00400, 0.01050, 25, TRUE),
('Enfermedad y Maternidad (Prestaciones en Dinero)', 0.00250, 0.00700, 25, TRUE),
('Invalidez y Vida', 0.00625, 0.01750, 25, TRUE),
('Guarderías y Prestaciones Sociales', 0.00000, 0.01000, 25, TRUE),
('Retiro', 0.00000, 0.02000, 25, TRUE),
('Cesantía en Edad Avanzada y Vejez', 0.01125, 0.00000, 25, FALSE), -- Progressive for employer
('Riesgo de Trabajo', 0.00000, 0.00500, 25, TRUE); -- Example rate, varies by company

-- Seed IMSS Employer Cesantía Progressive Brackets (2025)
INSERT INTO imss_employer_cesantia_brackets (fiscal_year_id, lower_bound_uma, upper_bound_uma, employer_percent) VALUES
((SELECT id FROM fiscal_years WHERE year = 2025), 1.000, 1.500, 0.03150),
((SELECT id FROM fiscal_years WHERE year = 2025), 1.501, 2.000, 0.03281),
((SELECT id FROM fiscal_years WHERE year = 2025), 2.001, 2.500, 0.03412),
((SELECT id FROM fiscal_years WHERE year = 2025), 2.501, 3.000, 0.03544),
((SELECT id FROM fiscal_years WHERE year = 2025), 3.001, 3.500, 0.03863),
((SELECT id FROM fiscal_years WHERE year = 2025), 3.501, 4.000, 0.04181),
((SELECT id FROM fiscal_years WHERE year = 2025), 4.001, 4.500, 0.04500),
((SELECT id FROM fiscal_years WHERE year = 2025), 4.501, 5.000, 0.04819),
((SELECT id FROM fiscal_years WHERE year = 2025), 5.001, 5.500, 0.05137),
((SELECT id FROM fiscal_years WHERE year = 2025), 5.501, 6.000, 0.05456),
((SELECT id FROM fiscal_years WHERE year = 2025), 6.001, 6.500, 0.05775),
((SELECT id FROM fiscal_years WHERE year = 2025), 6.501, 7.000, 0.06093),
((SELECT id FROM fiscal_years WHERE year = 2025), 7.001, 7.500, 0.06412),
((SELECT id FROM fiscal_years WHERE year = 2025), 7.501, 8.000, 0.06731),
((SELECT id FROM fiscal_years WHERE year = 2025), 8.001, 8.500, 0.07049),
((SELECT id FROM fiscal_years WHERE year = 2025), 8.501, 9.000, 0.07368),
((SELECT id FROM fiscal_years WHERE year = 2025), 9.001, 9.500, 0.07686),
((SELECT id FROM fiscal_years WHERE year = 2025), 9.501, 10.000, 0.08005),
((SELECT id FROM fiscal_years WHERE year = 2025), 10.001, 999.999, 0.08324); -- 25 UMAs and above

