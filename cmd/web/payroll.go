package main

import (
	"fmt"
	"math"

	"github.com/jcroyoaun/totalcompmx/internal/database"
)

// calculateRESICO performs RESICO regime calculation (flat rate, no IMSS, no subsidio)
func (app *application) calculateRESICO(
	monthlyIncome float64,
	unpaidVacationDays int,
	otherBenefits []OtherBenefit,
	exchangeRate float64,
	fiscalYear database.FiscalYear,
) (database.SalaryCalculation, error) {
	result := database.SalaryCalculation{
		GrossSalary:        monthlyIncome,
		UnpaidVacationDays: unpaidVacationDays,
	}

	// Get RESICO bracket
	resicoBracket, found, err := app.db.GetRESICOBracket(fiscalYear.ID, monthlyIncome)
	if err != nil {
		return result, err
	}
	if !found {
		return result, fmt.Errorf("no RESICO bracket found for income %.2f", monthlyIncome)
	}

	// RESICO: Apply flat rate to TOTAL income
	result.ISRTax = monthlyIncome * resicoBracket.ApplicableRate

	// RESICO has NO:
	// - IMSS (result.IMSSWorker = 0)
	// - Subsidio al Empleo (result.SubsidioEmpleo = 0)
	// - SBC (result.SBC = 0)
	
	// Calculate Net Salary
	result.NetSalary = monthlyIncome - result.ISRTax

	// Process Other Benefits (Otras prestaciones) - separate monthly and annual
	var otherBenefitsMonthlyNet float64
	var otherBenefitsAnnualNet float64
	
	for _, benefit := range otherBenefits {
		// Convert to MXN if needed
		benefitAmount := benefit.Amount
		if benefit.Currency == "USD" {
			benefitAmount = benefit.Amount * exchangeRate
		}
		
		benefitResult := database.OtherBenefitResult{
			Name:    benefit.Name,
			Amount:  benefitAmount,
			TaxFree: benefit.TaxFree,
			Cadence: benefit.Cadence,
		}
		
		if benefit.TaxFree {
			// Tax-free benefits
			benefitResult.ISR = 0
			benefitResult.Net = benefitAmount
		} else {
			// Taxable benefits - apply RESICO rate
			benefitResult.ISR = benefitAmount * resicoBracket.ApplicableRate
			benefitResult.Net = benefitAmount - benefitResult.ISR
		}
		
		result.OtherBenefits = append(result.OtherBenefits, benefitResult)
		
		// Add to monthly or annual based on cadence
		if benefit.Cadence == "annual" {
			otherBenefitsAnnualNet += benefitResult.Net
		} else {
			// Default to monthly
			otherBenefitsMonthlyNet += benefitResult.Net
		}
	}
	
	// Add monthly other benefits to monthly net
	result.OtherBenefitsMonthlyNet = otherBenefitsMonthlyNet
	result.NetSalary += otherBenefitsMonthlyNet

	// Calculate yearly totals (include annual benefits)
	result.YearlyGrossBase = monthlyIncome * 12
	result.YearlyGross = result.YearlyGrossBase
	
	// RESICO Unpaid Vacation Adjustment
	// Freelancers don't get paid when they don't work - this is "opportunity cost"
	if unpaidVacationDays > 0 {
		dailyRate := monthlyIncome / 30.4 // Average days per month
		result.UnpaidVacationLoss = dailyRate * float64(unpaidVacationDays)
		
		// Reduce yearly gross and net by the lost income
		result.YearlyGrossBase -= result.UnpaidVacationLoss
		result.YearlyGross -= result.UnpaidVacationLoss
	}
	
	result.YearlyNet = (result.NetSalary * 12) + otherBenefitsAnnualNet - result.UnpaidVacationLoss
	result.MonthlyAdjusted = result.YearlyNet / 12.0

	return result, nil
}

// calculateSalaryWithBenefits performs the full Mexican payroll calculation with benefits
func (app *application) calculateSalaryWithBenefits(
	grossMonthlySalary float64,
	hasAguinaldo bool, aguinaldoDays int,
	hasValesDespensa bool, valesDespensaAmount float64,
	hasPrimaVacacional bool, vacationDays int, primaVacacionalPercent float64,
	hasFondoAhorro bool, fondoAhorroPercent float64,
	hasInfonavitCredit bool,
	otherBenefits []OtherBenefit,
	exchangeRate float64,
	fiscalYear database.FiscalYear,
) (database.SalaryCalculation, error) {
	
	// Calculate monthly first
	result, err := app.calculateSalary(grossMonthlySalary, 1, fiscalYear)
	if err != nil {
		return result, err
	}
	
	// Apply Fondo de Ahorro monthly deduction
	if hasFondoAhorro {
		monthlyDeduction := grossMonthlySalary * (fondoAhorroPercent / 100.0)
		// Cap at 1.3 UMA annually / 12
		maxMonthlyDeduction := (fiscalYear.UMAAnnual * 1.3) / 12.0
		if monthlyDeduction > maxMonthlyDeduction {
			monthlyDeduction = maxMonthlyDeduction
		}
		result.FondoAhorroEmployee = monthlyDeduction
		result.NetSalary -= monthlyDeduction
	}
	
	// Add Vales de Despensa to monthly net (tax-free, max 1 UMA monthly)
	if hasValesDespensa {
		monthlyVales := valesDespensaAmount
		if monthlyVales > fiscalYear.UMAMonthly {
			monthlyVales = fiscalYear.UMAMonthly
		}
		result.ValesDespensaMonthly = monthlyVales
		result.NetSalary += monthlyVales
	}
	
	// Process Other Benefits (Otras prestaciones) - separate monthly and annual
	var otherBenefitsMonthlyNet float64
	var otherBenefitsAnnualNet float64
	
	for _, benefit := range otherBenefits {
		// Convert to MXN if needed
		benefitAmount := benefit.Amount
		if benefit.Currency == "USD" {
			benefitAmount = benefit.Amount * exchangeRate
		}
		
		benefitResult := database.OtherBenefitResult{
			Name:    benefit.Name,
			Amount:  benefitAmount,
			TaxFree: benefit.TaxFree,
			Cadence: benefit.Cadence,
		}
		
		if benefit.TaxFree {
			// Tax-free benefits
			benefitResult.ISR = 0
			benefitResult.Net = benefitAmount
		} else {
			// Taxable benefits
			isrBrackets, err := app.db.GetISRBrackets(fiscalYear.ID)
			if err != nil {
				return result, err
			}
			
			// Use Article 174 method for annual bonuses (considers base salary)
			// Use standard ISR for monthly benefits (isolated calculation)
			if benefit.Cadence == "annual" {
				benefitResult.ISR = calculateTaxArt174(grossMonthlySalary, benefitAmount, isrBrackets)
			} else {
				benefitResult.ISR = calculateISR(benefitAmount, isrBrackets)
			}
			benefitResult.Net = benefitAmount - benefitResult.ISR
		}
		
		result.OtherBenefits = append(result.OtherBenefits, benefitResult)
		
		// Add to monthly or annual based on cadence
		if benefit.Cadence == "annual" {
			otherBenefitsAnnualNet += benefitResult.Net
		} else {
			// Default to monthly
			otherBenefitsMonthlyNet += benefitResult.Net
		}
	}
	
	// Add monthly other benefits to monthly net
	result.OtherBenefitsMonthlyNet = otherBenefitsMonthlyNet
	result.NetSalary += otherBenefitsMonthlyNet
	
	// Calculate yearly components (paid once per year)
	dailySalary := grossMonthlySalary / 30.4
	
	// 1. Aguinaldo (subject to ISR with 30 UMA exemption per LISR Article 93, not subject to IMSS)
	if hasAguinaldo {
		result.AguinaldoGross = dailySalary * float64(aguinaldoDays)
		
		// LISR Article 93: Aguinaldo is exempt from ISR up to 30 UMAs
		exemptAmount := 30.0 * fiscalYear.UMADaily
		
		// Only the amount exceeding the exemption is taxable
		taxableBase := math.Max(0, result.AguinaldoGross-exemptAmount)
		
		// Calculate ISR on taxable base using Article 174 (progressive method)
		if taxableBase > 0 {
			isrBrackets, err := app.db.GetISRBrackets(fiscalYear.ID)
			if err != nil {
				return result, err
			}
			result.AguinaldoISR = calculateTaxArt174(grossMonthlySalary, taxableBase, isrBrackets)
		} else {
			result.AguinaldoISR = 0
		}
		
		result.AguinaldoNet = result.AguinaldoGross - result.AguinaldoISR
	}
	
	// 2. Prima Vacacional (subject to ISR with 15 UMA exemption per LISR Article 93, not subject to IMSS)
	if hasPrimaVacacional {
		vacationSalary := dailySalary * float64(vacationDays)
		result.PrimaVacacionalGross = vacationSalary * (primaVacacionalPercent / 100.0)
		
		// LISR Article 93: Prima Vacacional is exempt from ISR up to 15 UMAs
		exemptAmount := 15.0 * fiscalYear.UMADaily
		
		// Only the amount exceeding the exemption is taxable
		taxableBase := math.Max(0, result.PrimaVacacionalGross-exemptAmount)
		
		// Calculate ISR on taxable base using Article 174 (progressive method)
		if taxableBase > 0 {
			isrBrackets, err := app.db.GetISRBrackets(fiscalYear.ID)
			if err != nil {
				return result, err
			}
			result.PrimaVacacionalISR = calculateTaxArt174(grossMonthlySalary, taxableBase, isrBrackets)
		} else {
			result.PrimaVacacionalISR = 0
		}
		
		result.PrimaVacacionalNet = result.PrimaVacacionalGross - result.PrimaVacacionalISR
	}
	
	// 3. Fondo de Ahorro yearly return (company returns 2x employee contribution)
	if hasFondoAhorro {
		yearlyEmployeeContribution := result.FondoAhorroEmployee * 12
		// Company matches 100% (returns 2x what was deducted)
		result.FondoAhorroYearly = yearlyEmployeeContribution * 2
	}
	
	// 4. Infonavit Employer Contribution (Art 29, Ley Infonavit)
	// Employers pay 5% of SBC (already capped at 25 UMAs)
	// Paid bimonthly but shown as monthly equivalent
	// This is NON-LIQUID (goes to housing fund, not employee's pocket)
	monthlySBC := result.SBC * 30.4 // Daily SBC to Monthly
	result.InfonavitEmployerMonthly = monthlySBC * 0.05
	result.InfonavitEmployerAnnual = result.InfonavitEmployerMonthly * 12
	result.HasInfonavitCredit = hasInfonavitCredit // Flag to determine if it's mortgage payment or savings
	
	// 5. IMSS Employer Contributions (Non-liquid, part of total comp)
	imssEmployer, err := app.calculateIMSSEmployer(grossMonthlySalary, fiscalYear)
	if err != nil {
		return result, err
	}
	result.IMSSEmployerMonthly = imssEmployer
	result.IMSSEmployerAnnual = imssEmployer * 12
	
	// Calculate yearly totals
	result.YearlyGrossBase = grossMonthlySalary * 12 // Compensation bruta anual (solo salario)
	
	// YearlyGross includes:
	// - Base salary (12 months)
	// - Aguinaldo
	// - Prima Vacacional
	// - Infonavit Employer (12 months) - Non-liquid but part of total comp
	// - IMSS Employer (12 months) - Non-liquid but part of total comp
	result.YearlyGross = result.YearlyGrossBase + result.AguinaldoGross + result.PrimaVacacionalGross + 
		(result.InfonavitEmployerMonthly * 12) + (result.IMSSEmployerMonthly * 12)
	result.YearlyNet = (result.NetSalary * 12) + result.AguinaldoNet + result.PrimaVacacionalNet + result.FondoAhorroYearly + otherBenefitsAnnualNet
	result.MonthlyAdjusted = result.YearlyNet / 12.0
	
	return result, nil
}

// calculateSalary performs the full Mexican payroll calculation
func (app *application) calculateSalary(grossMonthlySalary float64, yearsOfService int, fiscalYear database.FiscalYear) (database.SalaryCalculation, error) {
	result := database.SalaryCalculation{
		GrossSalary: grossMonthlySalary,
	}

	// Calculate ISR Tax
	isrBrackets, err := app.db.GetISRBrackets(fiscalYear.ID)
	if err != nil {
		return result, err
	}

	result.ISRTax = calculateISR(grossMonthlySalary, isrBrackets)

	// Calculate Subsidio al Empleo (if applicable)
	if grossMonthlySalary <= fiscalYear.SubsidyThresholdMonthly {
		result.SubsidioEmpleo = grossMonthlySalary * fiscalYear.SubsidyFactor
	}

	// Calculate IMSS Worker contributions
	imssWorker, err := app.calculateIMSSWorker(grossMonthlySalary, fiscalYear)
	if err != nil {
		return result, err
	}
	result.IMSSWorker = imssWorker

	// Calculate SBC (Salario Base de Cotización)
	// For simplicity in MVP, using gross salary as base
	// In production, you'd calculate with aguinaldo, prima vacacional, etc.
	result.SBC = calculateSBC(grossMonthlySalary, yearsOfService, fiscalYear)

	// Calculate Net Salary
	// Net = Gross - ISR + Subsidio - IMSS - Other Deductions
	result.NetSalary = grossMonthlySalary - result.ISRTax + result.SubsidioEmpleo - result.IMSSWorker

	return result, nil
}

// calculateISR calculates the ISR tax based on progressive brackets
func calculateISR(grossSalary float64, brackets []database.ISRBracket) float64 {
	for _, bracket := range brackets {
		if grossSalary >= bracket.LowerLimit && grossSalary <= bracket.UpperLimit {
			surplus := grossSalary - bracket.LowerLimit
			isr := bracket.FixedFee + (surplus * bracket.SurplusPercent)
			return math.Round(isr*100) / 100 // Round to 2 decimals
		}
	}
	return 0
}

// calculateTaxArt174 calculates ISR on annual bonuses using Article 174 methodology
// This prevents under-taxation by considering that the base salary has already consumed lower brackets
func calculateTaxArt174(grossMonthlySalary, annualBonusAmount float64, brackets []database.ISRBracket) float64 {
	if annualBonusAmount <= 0 {
		return 0
	}
	
	// Step A: Convert bonus to daily rate, then to monthly equivalent
	// This represents what the bonus would be if spread over the year
	remuneracionMensual := (annualBonusAmount / 365.0) * 30.4
	
	// Step B: Calculate partial tax
	// Tax on (Salary + Monthly Share of Bonus)
	taxOnTotal := calculateISR(grossMonthlySalary+remuneracionMensual, brackets)
	
	// Tax on Salary alone
	taxOnSalary := calculateISR(grossMonthlySalary, brackets)
	
	// Tax attributable to the monthly share of the bonus
	taxOnShare := taxOnTotal - taxOnSalary
	
	// Step C: Calculate effective rate
	// This is the marginal rate at which this bonus should be taxed
	var effectiveRate float64
	if remuneracionMensual > 0 {
		effectiveRate = taxOnShare / remuneracionMensual
	}
	
	// Step D: Apply rate to full bonus
	taxToWithhold := annualBonusAmount * effectiveRate
	
	return math.Round(taxToWithhold*100) / 100 // Round to 2 decimals
}

// calculateIMSSWorker calculates the worker's IMSS contributions
func (app *application) calculateIMSSWorker(grossSalary float64, fiscalYear database.FiscalYear) (float64, error) {
	concepts, err := app.db.GetIMSSConcepts()
	if err != nil {
		return 0, err
	}

	dailySalary := grossSalary / 30.4 // Average days in a month
	var total float64

	for _, concept := range concepts {
		// Cap the base salary at 25 UMAs (or whatever the concept specifies)
		baseForCalculation := dailySalary
		if concept.BaseCapInUMAs > 0 {
			maxBase := float64(concept.BaseCapInUMAs) * fiscalYear.UMADaily
			if dailySalary > maxBase {
				baseForCalculation = maxBase
			}
		}

		// Calculate monthly contribution
		monthlyBase := baseForCalculation * 30.4
		contribution := monthlyBase * concept.WorkerPercent

		// Special handling for Cesantía (progressive)
		if !concept.IsFixedRate && concept.ConceptName == "Cesantía en Edad Avanzada y Vejez" {
			salaryInUMAs := dailySalary / fiscalYear.UMADaily
			_, found, err := app.db.GetCesantiaBracket(fiscalYear.ID, salaryInUMAs)
			if err != nil {
				return 0, err
			}
			if found {
				// Worker pays fixed 1.125%, but we already calculated it above
				// The progressive part is for employer
				contribution = monthlyBase * concept.WorkerPercent
			}
		}

		total += contribution
	}

	return math.Round(total*100) / 100, nil
}

// calculateIMSSEmployer calculates the employer's IMSS contributions
// This is NON-LIQUID compensation (doesn't go to employee's pocket)
func (app *application) calculateIMSSEmployer(grossSalary float64, fiscalYear database.FiscalYear) (float64, error) {
	concepts, err := app.db.GetIMSSConcepts()
	if err != nil {
		return 0, err
	}

	dailySalary := grossSalary / 30.4 // Average days in a month
	var total float64

	for _, concept := range concepts {
		// Cap the base salary at 25 UMAs (or whatever the concept specifies)
		baseForCalculation := dailySalary
		if concept.BaseCapInUMAs > 0 {
			maxBase := float64(concept.BaseCapInUMAs) * fiscalYear.UMADaily
			if dailySalary > maxBase {
				baseForCalculation = maxBase
			}
		}

		// Calculate monthly contribution
		monthlyBase := baseForCalculation * 30.4
		contribution := monthlyBase * concept.EmployerPercent

		// Special handling for Cesantía (progressive for employer)
		if !concept.IsFixedRate && concept.ConceptName == "Cesantía en Edad Avanzada y Vejez" {
			salaryInUMAs := dailySalary / fiscalYear.UMADaily
			bracket, found, err := app.db.GetCesantiaBracket(fiscalYear.ID, salaryInUMAs)
			if err != nil {
				return 0, err
			}
			if found {
				// Employer pays progressive rate based on bracket
				contribution = monthlyBase * bracket.EmployerPercent
			}
		}

		total += contribution
	}

	return math.Round(total*100) / 100, nil
}

// calculateSBC calculates the Salario Base de Cotización
// This is a simplified version for MVP
func calculateSBC(grossMonthlySalary float64, yearsOfService int, fiscalYear database.FiscalYear) float64 {
	// Simplified: In reality, you'd need to calculate with aguinaldo, prima vacacional, etc.
	// For MVP, we'll use a basic factor
	
	// Default integration factor (aguinaldo 15 days + prima vac 12 days * 25% = 18 days / 365)
	integrationFactor := 1.0493 // Approximately 4.93% additional for benefits

	// If years of service > 0, adjust slightly (simplified)
	if yearsOfService > 0 {
		// Each year adds slight increase due to more vacation days
		additionalFactor := float64(yearsOfService) * 0.001
		integrationFactor += additionalFactor
	}

	dailySalary := grossMonthlySalary / 30.4
	sbc := dailySalary * integrationFactor

	// Cap at 25 UMAs
	maxSBC := 25 * fiscalYear.UMADaily
	if sbc > maxSBC {
		sbc = maxSBC
	}

	return math.Round(sbc*100) / 100
}

