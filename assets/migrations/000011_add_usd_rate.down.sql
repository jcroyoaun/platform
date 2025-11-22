-- Rollback USD/MXN exchange rate column

ALTER TABLE fiscal_years 
DROP COLUMN IF EXISTS usd_mxn_rate;

