package database

import (
	"context"
	"database/sql"
)

type FiscalYear struct {
	ID                        int
	Year                      int
	UMADaily                  float64
	UMAMonthly                float64
	UMAAnnual                 float64
	UMIValue                  float64
	SMGGeneral                float64
	SMGBorder                 float64
	SubsidyFactor             float64
	SubsidyThresholdMonthly   float64
	FALegalCapUMAFactor       float64
	FALegalMaxPercentage      float64
	PantryVouchersUMACap      float64
	USDMXNRate                float64 // Exchange rate USD/MXN
}

type ISRBracket struct {
	LowerLimit     float64
	UpperLimit     float64
	FixedFee       float64
	SurplusPercent float64
}

type IMSSConcept struct {
	ConceptName     string
	WorkerPercent   float64
	EmployerPercent float64
	BaseCapInUMAs   int
	IsFixedRate     bool
}

type CesantiaBracket struct {
	LowerBoundUMA   float64
	UpperBoundUMA   float64
	EmployerPercent float64
}

type RESICOBracket struct {
	UpperLimit      float64
	ApplicableRate  float64
}

type SalaryCalculation struct {
	// Monthly
	GrossSalary             float64
	ISRTax                  float64
	SubsidioEmpleo          float64
	IMSSWorker              float64
	FondoAhorroEmployee     float64
	InfonavitDiscount       float64
	ValesDespensaMonthly    float64 // Added to monthly net
	OtherBenefitsMonthlyNet float64 // Monthly otras prestaciones added to net
	NetSalary               float64
	SBC                     float64 // Salario Base de Cotización
	
	// Yearly Components (paid once a year)
	AguinaldoGross       float64
	AguinaldoISR         float64
	AguinaldoNet         float64
	PrimaVacacionalGross float64
	PrimaVacacionalISR   float64
	PrimaVacacionalNet   float64
	FondoAhorroYearly    float64 // What company returns (2x employee contribution)
	
	// Employer Contributions (Non-Liquid, Total Comp only)
	InfonavitEmployerMonthly float64 // 5% of capped SBC, paid bimonthly but shown as monthly
	InfonavitEmployerAnnual  float64 // Infonavit x 12
	IMSSEmployerMonthly      float64 // Total employer IMSS contributions
	IMSSEmployerAnnual       float64 // IMSS x 12
	HasInfonavitCredit       bool    // True if employee has an Infonavit mortgage
	
	// Totals
	YearlyGrossBase     float64 // Salary * 12 (no benefits)
	YearlyGross         float64 // Total including all benefits
	YearlyNet           float64
	MonthlyAdjusted     float64 // Monthly net + (yearly benefits / 12)
	
	// RESICO Specific
	UnpaidVacationDays  int     // RESICO only: days off without pay
	UnpaidVacationLoss  float64 // RESICO only: income lost due to unpaid days off
	
	// Other Benefits
	OtherBenefits []OtherBenefitResult
}

type OtherBenefitResult struct {
	Name    string
	Amount  float64
	TaxFree bool
	ISR     float64
	Net     float64
	Cadence string // "monthly" or "annual"
}

// GetActiveFiscalYear retrieves the active fiscal year configuration
func (db *DB) GetActiveFiscalYear() (FiscalYear, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT id, year, uma_daily, uma_monthly, uma_annual, umi_value,
		       smg_general, smg_border, subsidy_factor, subsidy_threshold_monthly,
		       fa_legal_cap_uma_factor, fa_legal_max_percentage, pantry_vouchers_uma_cap,
		       COALESCE(usd_mxn_rate, 20.00) as usd_mxn_rate
		FROM fiscal_years
		WHERE is_active = true
		LIMIT 1`

	var fy FiscalYear
	err := db.QueryRowContext(ctx, query).Scan(
		&fy.ID, &fy.Year, &fy.UMADaily, &fy.UMAMonthly, &fy.UMAAnnual, &fy.UMIValue,
		&fy.SMGGeneral, &fy.SMGBorder, &fy.SubsidyFactor, &fy.SubsidyThresholdMonthly,
		&fy.FALegalCapUMAFactor, &fy.FALegalMaxPercentage, &fy.PantryVouchersUMACap,
		&fy.USDMXNRate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return FiscalYear{}, false, nil
		}
		return FiscalYear{}, false, err
	}

	return fy, true, nil
}

// GetISRBrackets retrieves all ISR tax brackets for monthly calculations
func (db *DB) GetISRBrackets(fiscalYearID int) ([]ISRBracket, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT lower_limit, upper_limit, fixed_fee, surplus_percent
		FROM isr_brackets
		WHERE fiscal_year_id = $1 AND periodicity = 'MONTHLY'
		ORDER BY lower_limit ASC`

	rows, err := db.QueryContext(ctx, query, fiscalYearID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brackets []ISRBracket
	for rows.Next() {
		var b ISRBracket
		err := rows.Scan(&b.LowerLimit, &b.UpperLimit, &b.FixedFee, &b.SurplusPercent)
		if err != nil {
			return nil, err
		}
		brackets = append(brackets, b)
	}

	return brackets, rows.Err()
}

// GetIMSSConcepts retrieves all IMSS contribution concepts
func (db *DB) GetIMSSConcepts() ([]IMSSConcept, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT concept_name, worker_percent, employer_percent, base_cap_in_umas, is_fixed_rate
		FROM imss_concepts`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var concepts []IMSSConcept
	for rows.Next() {
		var c IMSSConcept
		err := rows.Scan(&c.ConceptName, &c.WorkerPercent, &c.EmployerPercent, &c.BaseCapInUMAs, &c.IsFixedRate)
		if err != nil {
			return nil, err
		}
		concepts = append(concepts, c)
	}

	return concepts, rows.Err()
}

// GetCesantiaBracket retrieves the Cesantía bracket for a given salary
func (db *DB) GetCesantiaBracket(fiscalYearID int, salaryInUMAs float64) (CesantiaBracket, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT lower_bound_uma, upper_bound_uma, employer_percent
		FROM imss_employer_cesantia_brackets
		WHERE fiscal_year_id = $1 
		  AND $2 >= lower_bound_uma 
		  AND $2 <= upper_bound_uma
		LIMIT 1`

	var cb CesantiaBracket
	err := db.QueryRowContext(ctx, query, fiscalYearID, salaryInUMAs).Scan(
		&cb.LowerBoundUMA, &cb.UpperBoundUMA, &cb.EmployerPercent,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return CesantiaBracket{}, false, nil
		}
		return CesantiaBracket{}, false, err
	}

	return cb, true, nil
}

// GetRESICOBracket retrieves the RESICO bracket for a given monthly income
func (db *DB) GetRESICOBracket(fiscalYearID int, monthlyIncome float64) (RESICOBracket, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT upper_limit, applicable_rate
		FROM resico_brackets
		WHERE fiscal_year_id = $1 
		  AND periodicity = 'MONTHLY'
		  AND $2 <= upper_limit
		ORDER BY upper_limit ASC
		LIMIT 1`

	var rb RESICOBracket
	err := db.QueryRowContext(ctx, query, fiscalYearID, monthlyIncome).Scan(
		&rb.UpperLimit, &rb.ApplicableRate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return RESICOBracket{}, false, nil
		}
		return RESICOBracket{}, false, err
	}

	return rb, true, nil
}

// UpdateExchangeRate updates the USD/MXN exchange rate for the active fiscal year
func (db *DB) UpdateExchangeRate(rate float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		UPDATE fiscal_years 
		SET usd_mxn_rate = $1
		WHERE is_active = true`

	result, err := db.ExecContext(ctx, query, rate)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// UpdateUMA updates the UMA values for the active fiscal year
func (db *DB) UpdateUMA(annual, monthly, daily float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		UPDATE fiscal_years 
		SET uma_annual = $1,
		    uma_monthly = $2,
		    uma_daily = $3
		WHERE is_active = true`

	result, err := db.ExecContext(ctx, query, annual, monthly, daily)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

