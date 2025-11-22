-- Add USD/MXN exchange rate column to fiscal_years table

ALTER TABLE fiscal_years 
ADD COLUMN IF NOT EXISTS usd_mxn_rate NUMERIC(10, 4) DEFAULT 20.00;

-- Add comment
COMMENT ON COLUMN fiscal_years.usd_mxn_rate IS 'Current USD/MXN exchange rate from Banxico (updated daily)';

-- Update existing records with default value
UPDATE fiscal_years 
SET usd_mxn_rate = 20.00 
WHERE usd_mxn_rate IS NULL;

