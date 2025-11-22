package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jcroyoaun/totalcompmx/internal/database"
	"github.com/jcroyoaun/totalcompmx/internal/equity"
	"github.com/jcroyoaun/totalcompmx/internal/password"
	"github.com/jcroyoaun/totalcompmx/internal/pdf"
	"github.com/jcroyoaun/totalcompmx/internal/request"
	"github.com/jcroyoaun/totalcompmx/internal/response"
	"github.com/jcroyoaun/totalcompmx/internal/token"
	"github.com/jcroyoaun/totalcompmx/internal/validator"
)

type OtherBenefit struct {
	Name         string
	Amount       float64
	TaxFree      bool
	Currency     string
	Cadence      string // monthly, annual, etc.
	IsPercentage bool   // true if Amount is a percentage of gross annual salary
}

type PackageResult struct {
	PackageName     string
	*database.SalaryCalculation
	EquityConfig    *equity.EquityConfig
	EquitySchedule  []equity.YearlyEquity
}

type PackageInput struct {
	Name                    string
	Regime                  string
	Currency                string
	ExchangeRate            string
	PaymentFrequency        string
	HoursPerWeek            string
	GrossMonthlySalary      string
	HasAguinaldo            bool
	AguinaldoDays           string
	HasValesDespensa        bool
	ValesDespensaAmount     string
	HasPrimaVacacional      bool
	VacationDays            string
	PrimaVacacionalPercent  string
	HasFondoAhorro          bool
	FondoAhorroPercent      string
	UnpaidVacationDays      string // RESICO only: days off without pay
	OtherBenefits           []OtherBenefit
	// Equity fields
	HasEquity               bool
	InitialEquityUSD        string
	HasRefreshers           bool
	RefresherMinUSD         string
	RefresherMaxUSD         string
}

func (app *application) clearSession(w http.ResponseWriter, r *http.Request) {
	// Clear all session data
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	
	// Remove all calculator-related session data
	app.sessionManager.Remove(r.Context(), "packageInputs")
	app.sessionManager.Remove(r.Context(), "comparisonResults")
	app.sessionManager.Remove(r.Context(), "bestPackage")
	app.sessionManager.Remove(r.Context(), "fiscalYear")
	
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	var form struct {
		PackageNames []string `form:"-"`
		Validator    validator.Validator `form:"-"`
	}

	switch r.Method {
	case http.MethodGet:
		data := app.newTemplateData(r)
		
		// Always fetch FiscalYear for exchange rate display
		fiscalYear, found, err := app.db.GetActiveFiscalYear()
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		if found {
			data["FiscalYear"] = fiscalYear
		}
		
		// Check if we have comparison results in session (from POST-Redirect-GET)
		if app.sessionManager.Exists(r.Context(), "comparisonResults") {
			// Load from session
			data["PackageInputs"] = app.sessionManager.Get(r.Context(), "packageInputs")
			data["Results"] = app.sessionManager.Get(r.Context(), "comparisonResults")
			data["BestPackage"] = app.sessionManager.Get(r.Context(), "bestPackage")
			// FiscalYear already set above, but override if session has it
			if sessionFiscalYear := app.sessionManager.Get(r.Context(), "fiscalYear"); sessionFiscalYear != nil {
				data["FiscalYear"] = sessionFiscalYear
			}
			
			// Keep session data for PDF export - don't clear it here
			// Session will be cleared when user submits a new comparison
		} else {
			// No results yet, show empty form
			form.PackageNames = []string{"Paquete 1", "Paquete 2"}
			data["Form"] = form
		}

		err = response.Page(w, http.StatusOK, data, "pages/home.tmpl")
		if err != nil {
			app.serverError(w, r, err)
		}

	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			app.badRequest(w, r, err)
			return
		}

		// Get fiscal year configuration
		fiscalYear, found, err := app.db.GetActiveFiscalYear()
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		if !found {
			app.serverError(w, r, fmt.Errorf("no active fiscal year found"))
			return
		}

		// Parse arrays from form
		packageNames := r.Form["PackageName[]"]
		regimes := r.Form["Regime[]"]
		salariesStr := r.Form["GrossMonthlySalary[]"]
		currencies := r.Form["Currency[]"]
		exchangeRatesStr := r.Form["ExchangeRate[]"]
		paymentFrequencies := r.Form["PaymentFrequency[]"]
		hoursPerWeekStr := r.Form["HoursPerWeek[]"]
		hasAguinaldo := r.Form["HasAguinaldo[]"]
		aguinaldoDaysStr := r.Form["AguinaldoDays[]"]
		hasValesDespensa := r.Form["HasValesDespensa[]"]
		valesDespensaAmountStr := r.Form["ValesDespensaAmount[]"]
		hasPrimaVacacional := r.Form["HasPrimaVacacional[]"]
		vacationDaysStr := r.Form["VacationDays[]"]
		primaVacacionalPercentStr := r.Form["PrimaVacacionalPercent[]"]
		hasFondoAhorro := r.Form["HasFondoAhorro[]"]
		fondoAhorroPercentStr := r.Form["FondoAhorroPercent[]"]
		hasInfonavitCredit := r.Form["HasInfonavitCredit[]"]
		unpaidVacationDaysStr := r.Form["UnpaidVacationDays[]"]
		
		// Equity form data
		hasEquity := r.Form["HasEquity[]"]
		initialEquityUSDStr := r.Form["InitialEquityUSD[]"]
		hasRefreshers := r.Form["HasRefreshers[]"]
		refresherMinUSDStr := r.Form["RefresherMinUSD[]"]
		refresherMaxUSDStr := r.Form["RefresherMaxUSD[]"]

		var results []PackageResult
		var bestPackage *PackageResult
		var packageInputs []PackageInput

		// Process each package (default to 2)
		numPackages := len(salariesStr)
		if numPackages == 0 {
			numPackages = 2
		}
		
		// Validate at least one package has a valid salary
		hasValidPackage := false
		for i := 0; i < numPackages; i++ {
			if i < len(salariesStr) && salariesStr[i] != "" {
				salary := 0.0
				fmt.Sscanf(salariesStr[i], "%f", &salary)
				if salary > 0 {
					hasValidPackage = true
					break
				}
			}
		}
		
		if !hasValidPackage {
			form.Validator.AddFieldError("GrossMonthlySalary", "Debes ingresar al menos un salario válido para comparar")
			
			// Restore form with error
			data := app.newTemplateData(r)
			data["Form"] = form
			err = response.Page(w, http.StatusUnprocessableEntity, data, "pages/home.tmpl")
			if err != nil {
				app.serverError(w, r, err)
			}
			return
		}

		for i := 0; i < numPackages; i++ {
			// Parse salary
			salaryStr := ""
			if i < len(salariesStr) {
				salaryStr = salariesStr[i]
			}
			salary := 0.0
			if salaryStr != "" {
				fmt.Sscanf(salaryStr, "%f", &salary)
			}

			if salary <= 0 {
				continue // Skip invalid packages
			}

			// Determine regime
			regime := "sueldos_salarios"
			if i < len(regimes) {
				regime = regimes[i]
			}

			// Currency conversion (USD -> MXN if needed)
			currency := "MXN"
			exchangeRate := 20.0 // Default exchange rate
			if i < len(currencies) {
				currency = currencies[i]
			}
			if currency == "USD" {
				if i < len(exchangeRatesStr) && exchangeRatesStr[i] != "" {
					fmt.Sscanf(exchangeRatesStr[i], "%f", &exchangeRate)
				}
				salary = salary * exchangeRate // Convert USD to MXN
			}

			// Payment frequency conversion (convert to monthly if needed)
			paymentFreq := "monthly"
			if i < len(paymentFrequencies) && paymentFrequencies[i] != "" {
				paymentFreq = paymentFrequencies[i]
			}
			
			switch paymentFreq {
			case "hourly":
				hoursPerWeek := 40.0 // Default
				if i < len(hoursPerWeekStr) && hoursPerWeekStr[i] != "" {
					fmt.Sscanf(hoursPerWeekStr[i], "%f", &hoursPerWeek)
				}
				// Convert hourly to monthly: rate * hours/week * 4.33 weeks/month
				salary = salary * hoursPerWeek * 4.33
			case "daily":
				// Convert daily to monthly: daily * 30 days/month
				salary = salary * 30
			case "weekly":
				// Convert weekly to monthly: weekly * 4.33 weeks/month
				salary = salary * 4.33
			case "biweekly":
				// Convert biweekly to monthly: biweekly * 2.17 (26 periods / 12 months)
				salary = salary * 2.17
			case "monthly":
				// Already monthly, no conversion needed
			}

			// Now salary is in MXN monthly

			// Parse benefits (only for Sueldos y Salarios)
			hasAguin := false
			aguinDays := 15
			hasVales := false
			valesAmount := 0.0
			hasPrima := false
			vacaDays := 12
			primaPercent := 25.0
			hasFondo := false
			fondoPercent := 13.0
			hasInfonavit := false
			unpaidVacationDays := 0

			if regime == "sueldos_salarios" {
				// Check if this package has aguinaldo
				for _, val := range hasAguinaldo {
					if val == fmt.Sprintf("%d", i) {
						hasAguin = true
						break
					}
				}
				if hasAguin && i < len(aguinaldoDaysStr) {
					fmt.Sscanf(aguinaldoDaysStr[i], "%d", &aguinDays)
				}

				// Check vales
				for _, val := range hasValesDespensa {
					if val == fmt.Sprintf("%d", i) {
						hasVales = true
						break
					}
				}
				if hasVales && i < len(valesDespensaAmountStr) {
					fmt.Sscanf(valesDespensaAmountStr[i], "%f", &valesAmount)
				}

				// Check prima
				for _, val := range hasPrimaVacacional {
					if val == fmt.Sprintf("%d", i) {
						hasPrima = true
						break
					}
				}
				if hasPrima {
					if i < len(vacationDaysStr) {
						fmt.Sscanf(vacationDaysStr[i], "%d", &vacaDays)
					}
					if i < len(primaVacacionalPercentStr) {
						fmt.Sscanf(primaVacacionalPercentStr[i], "%f", &primaPercent)
					}
				}

				// Check fondo
				for _, val := range hasFondoAhorro {
					if val == fmt.Sprintf("%d", i) {
						hasFondo = true
						break
					}
				}
				if hasFondo && i < len(fondoAhorroPercentStr) {
					fmt.Sscanf(fondoAhorroPercentStr[i], "%f", &fondoPercent)
				}
				
				// Check if this package has Infonavit credit
				for _, val := range hasInfonavitCredit {
					if val == fmt.Sprintf("%d", i) {
						hasInfonavit = true
						break
					}
				}
			} else if regime == "resico" {
				// Parse unpaid vacation days for RESICO
				if i < len(unpaidVacationDaysStr) && unpaidVacationDaysStr[i] != "" {
					fmt.Sscanf(unpaidVacationDaysStr[i], "%d", &unpaidVacationDays)
				}
			}

			// Parse "Otras prestaciones" for this package
			otherBenefits := []OtherBenefit{}
			otherNamesKey := fmt.Sprintf("OtherBenefitName-%d[]", i)
			otherAmountsKey := fmt.Sprintf("OtherBenefitAmount-%d[]", i)
			otherTaxFreeKey := fmt.Sprintf("OtherBenefitTaxFree-%d[]", i)
			otherCurrencyKey := fmt.Sprintf("OtherBenefitCurrency-%d[]", i)
			otherCadenceKey := fmt.Sprintf("OtherBenefitCadence-%d[]", i)
			otherTypeKey := fmt.Sprintf("OtherBenefitType-%d[]", i)
			
			otherNames := r.Form[otherNamesKey]
			otherAmounts := r.Form[otherAmountsKey]
			otherTaxFree := r.Form[otherTaxFreeKey]
			otherCurrency := r.Form[otherCurrencyKey]
			otherCadence := r.Form[otherCadenceKey]
			otherTypes := r.Form[otherTypeKey]
			
			// Build otherBenefits slice
			for j := 0; j < len(otherNames); j++ {
				if j >= len(otherAmounts) {
					break
				}
				name := otherNames[j]
				amountStr := otherAmounts[j]
				amount := 0.0
				fmt.Sscanf(amountStr, "%f", &amount)
				
				if name == "" || amount <= 0 {
					continue
				}
				
				// Check if this benefit is tax-free
				isTaxFree := false
				checkVal := fmt.Sprintf("%d", j+1)
				for _, val := range otherTaxFree {
					if val == checkVal {
						isTaxFree = true
						break
					}
				}
				
				// Get currency and cadence
				benefitCurrency := "MXN"
				if j < len(otherCurrency) {
					benefitCurrency = otherCurrency[j]
				}
				
				benefitCadence := "monthly"
				if j < len(otherCadence) {
					benefitCadence = otherCadence[j]
				}
				
				// Check if this is a percentage benefit
				isPercentage := false
				if j < len(otherTypes) && otherTypes[j] == "percentage" {
					isPercentage = true
					// For percentage, force annual cadence
					benefitCadence = "annual"
				}
				
				otherBenefits = append(otherBenefits, OtherBenefit{
					Name:         name,
					Amount:       amount,
					TaxFree:      isTaxFree,
					Currency:     benefitCurrency,
					Cadence:      benefitCadence,
					IsPercentage: isPercentage,
				})
			}

			// Calculate this package based on regime
			var result database.SalaryCalculation
			var err error
			
			if regime == "resico" {
				// RESICO: Simple flat rate calculation, no IMSS, no subsidio
				result, err = app.calculateRESICO(salary, unpaidVacationDays, otherBenefits, exchangeRate, fiscalYear)
			} else {
				// Sueldos y Salarios: Full calculation with benefits, IMSS, etc.
				result, err = app.calculateSalaryWithBenefits(
					salary,
					hasAguin, aguinDays,
					hasVales, valesAmount,
					hasPrima, vacaDays, primaPercent,
					hasFondo, fondoPercent,
					hasInfonavit,
					otherBenefits,
					exchangeRate,
					fiscalYear,
				)
			}
			
			if err != nil {
				app.serverError(w, r, err)
				return
			}

			packageName := fmt.Sprintf("Paquete %d", i+1)
			if i < len(packageNames) && packageNames[i] != "" {
				packageName = packageNames[i]
			}

			// Calculate equity if provided
			var equityConfig *equity.EquityConfig
			var equitySchedule []equity.YearlyEquity
			
			if i < len(initialEquityUSDStr) && initialEquityUSDStr[i] != "" {
				initialEquity := 0.0
				fmt.Sscanf(initialEquityUSDStr[i], "%f", &initialEquity)
				
				if initialEquity > 0 {
					refresherMin := 0.0
					refresherMax := 0.0
					hasRefresh := false
					
					if i < len(hasRefreshers) {
						for _, val := range hasRefreshers {
							if val == fmt.Sprintf("%d", i) {
								hasRefresh = true
								break
							}
						}
					}
					
					if hasRefresh && i < len(refresherMinUSDStr) && i < len(refresherMaxUSDStr) {
						fmt.Sscanf(refresherMinUSDStr[i], "%f", &refresherMin)
						fmt.Sscanf(refresherMaxUSDStr[i], "%f", &refresherMax)
						
						// Validate min <= max
						if refresherMin > refresherMax {
							refresherMin, refresherMax = refresherMax, refresherMin
						}
					}
					
					equityConfig = &equity.EquityConfig{
						InitialGrantUSD: initialEquity,
						HasRefreshers:   hasRefresh && refresherMin > 0 && refresherMax > 0,
						RefresherMinUSD: refresherMin,
						RefresherMaxUSD: refresherMax,
						VestingYears:    4,
						ExchangeRate:    fiscalYear.USDMXNRate,
					}
					
					equitySchedule = equity.CalculateEquitySchedule(*equityConfig, 4)
				}
			}

			packageResult := PackageResult{
				PackageName:       packageName,
				SalaryCalculation: &result,
				EquityConfig:      equityConfig,
				EquitySchedule:    equitySchedule,
			}

			results = append(results, packageResult)

			// Capture input values to preserve them
			exchangeRateStr := ""
			if i < len(exchangeRatesStr) && exchangeRatesStr[i] != "" {
				exchangeRateStr = exchangeRatesStr[i]
			}
			
			hoursStr := ""
			if i < len(hoursPerWeekStr) && hoursPerWeekStr[i] != "" {
				hoursStr = hoursPerWeekStr[i]
			}
			
			// Parse equity fields
			hasEquityChecked := false
			if i < len(hasEquity) {
				for _, val := range hasEquity {
					if val == fmt.Sprintf("%d", i) {
						hasEquityChecked = true
						break
					}
				}
			}
			
			initialEquityUSDVal := ""
			if i < len(initialEquityUSDStr) && initialEquityUSDStr[i] != "" {
				initialEquityUSDVal = initialEquityUSDStr[i]
			}
			
			hasRefresh := false
			if i < len(hasRefreshers) {
				for _, val := range hasRefreshers {
					if val == fmt.Sprintf("%d", i) {
						hasRefresh = true
						break
					}
				}
			}
			
			refresherMinVal := ""
			if i < len(refresherMinUSDStr) && refresherMinUSDStr[i] != "" {
				refresherMinVal = refresherMinUSDStr[i]
			}
			
			refresherMaxVal := ""
			if i < len(refresherMaxUSDStr) && refresherMaxUSDStr[i] != "" {
				refresherMaxVal = refresherMaxUSDStr[i]
			}
			
			packageInput := PackageInput{
				Name:                   packageName,
				Regime:                 regime,
				Currency:               currency,
				ExchangeRate:           exchangeRateStr,
				PaymentFrequency:       paymentFreq,
				HoursPerWeek:           hoursStr,
				GrossMonthlySalary:     salaryStr,
				HasAguinaldo:           hasAguin,
				AguinaldoDays:          fmt.Sprintf("%d", aguinDays),
				HasValesDespensa:       hasVales,
				ValesDespensaAmount:    fmt.Sprintf("%.2f", valesAmount),
				HasPrimaVacacional:     hasPrima,
				VacationDays:           fmt.Sprintf("%d", vacaDays),
				PrimaVacacionalPercent: fmt.Sprintf("%.2f", primaPercent),
				HasFondoAhorro:         hasFondo,
				FondoAhorroPercent:     fmt.Sprintf("%.2f", fondoPercent),
				UnpaidVacationDays:     fmt.Sprintf("%d", unpaidVacationDays),
				OtherBenefits:          otherBenefits,
				HasEquity:              hasEquityChecked,
				InitialEquityUSD:       initialEquityUSDVal,
				HasRefreshers:          hasRefresh,
				RefresherMinUSD:        refresherMinVal,
				RefresherMaxUSD:        refresherMaxVal,
			}
			
			packageInputs = append(packageInputs, packageInput)

			// Determine best package (highest yearly net)
			if bestPackage == nil || result.YearlyNet > bestPackage.SalaryCalculation.YearlyNet {
				bestPackage = &packageResult
			}
		}

		// Store results in session
		app.sessionManager.Put(r.Context(), "packageInputs", packageInputs)
		app.sessionManager.Put(r.Context(), "comparisonResults", results)
		app.sessionManager.Put(r.Context(), "bestPackage", bestPackage)
		app.sessionManager.Put(r.Context(), "fiscalYear", fiscalYear)

		// Redirect to GET to prevent form resubmission on refresh (POST-Redirect-GET pattern)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (app *application) signup(w http.ResponseWriter, r *http.Request) {
	var form struct {
		Email     string              `form:"Email"`
		Password  string              `form:"Password"`
		Validator validator.Validator `form:"-"`
	}

	switch r.Method {
	case http.MethodGet:
		data := app.newTemplateData(r)
		data["Form"] = form

		err := response.Page(w, http.StatusOK, data, "pages/signup.tmpl")
		if err != nil {
			app.serverError(w, r, err)
		}

	case http.MethodPost:
		err := request.DecodePostForm(r, &form)
		if err != nil {
			app.badRequest(w, r, err)
			return
		}

		_, found, err := app.db.GetUserByEmail(form.Email)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		form.Validator.CheckField(form.Email != "", "Email", "Email is required")
		form.Validator.CheckField(validator.Matches(form.Email, validator.RgxEmail), "Email", "Must be a valid email address")
		form.Validator.CheckField(!found, "Email", "Email is already in use")

		form.Validator.CheckField(form.Password != "", "Password", "Password is required")
		form.Validator.CheckField(len(form.Password) >= 8, "Password", "Password is too short")
		form.Validator.CheckField(len(form.Password) <= 72, "Password", "Password is too long")
		form.Validator.CheckField(validator.NotIn(form.Password, password.CommonPasswords...), "Password", "Password is too common")

		if form.Validator.HasErrors() {
			data := app.newTemplateData(r)
			data["Form"] = form

			err := response.Page(w, http.StatusUnprocessableEntity, data, "pages/signup.tmpl")
			if err != nil {
				app.serverError(w, r, err)
			}
			return
		}

		hashedPassword, err := password.Hash(form.Password)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		id, err := app.db.InsertUser(form.Email, hashedPassword)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		err = app.sessionManager.RenewToken(r.Context())
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

		// Generate verification token
		plaintextToken := token.New()
		hashedToken := token.Hash(plaintextToken)

		// Store verification token in database
		err = app.db.InsertEmailVerificationToken(id, hashedToken)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		// Send welcome email with verification link in background
		app.backgroundTask(r, func() error {
			data := app.newEmailData()
			data["Email"] = form.Email
			data["VerificationToken"] = plaintextToken
			return app.mailer.Send(form.Email, data, "welcome.tmpl")
		})

		http.Redirect(w, r, "/account/developer", http.StatusSeeOther)
	}
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	var form struct {
		Email     string              `form:"Email"`
		Password  string              `form:"Password"`
		Validator validator.Validator `form:"-"`
	}

	switch r.Method {
	case http.MethodGet:
		data := app.newTemplateData(r)
		data["Form"] = form

		err := response.Page(w, http.StatusOK, data, "pages/login.tmpl")
		if err != nil {
			app.serverError(w, r, err)
		}

	case http.MethodPost:
		err := request.DecodePostForm(r, &form)
		if err != nil {
			app.badRequest(w, r, err)
			return
		}

		user, found, err := app.db.GetUserByEmail(form.Email)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		form.Validator.CheckField(form.Email != "", "Email", "Email is required")
		form.Validator.CheckField(found, "Email", "Email address could not be found")

		if found {
			passwordMatches, err := password.Matches(form.Password, user.HashedPassword)
			if err != nil {
				app.serverError(w, r, err)
				return
			}

			form.Validator.CheckField(form.Password != "", "Password", "Password is required")
			form.Validator.CheckField(passwordMatches, "Password", "Password is incorrect")
		}

		if form.Validator.HasErrors() {
			data := app.newTemplateData(r)
			data["Form"] = form

			err := response.Page(w, http.StatusUnprocessableEntity, data, "pages/login.tmpl")
			if err != nil {
				app.serverError(w, r, err)
			}
			return
		}

		err = app.sessionManager.RenewToken(r.Context())
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		app.sessionManager.Put(r.Context(), "authenticatedUserID", user.ID)

		redirectPath := app.sessionManager.PopString(r.Context(), "redirectPathAfterLogin")
		if redirectPath != "" {
			http.Redirect(w, r, redirectPath, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/account/developer", http.StatusSeeOther)
	}
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) forgottenPassword(w http.ResponseWriter, r *http.Request) {
	var form struct {
		Email     string              `form:"Email"`
		Validator validator.Validator `form:"-"`
	}

	switch r.Method {
	case http.MethodGet:
		data := app.newTemplateData(r)
		data["Form"] = form

		err := response.Page(w, http.StatusOK, data, "pages/forgotten-password.tmpl")
		if err != nil {
			app.serverError(w, r, err)
		}

	case http.MethodPost:
		err := request.DecodePostForm(r, &form)
		if err != nil {
			app.badRequest(w, r, err)
			return
		}

		user, found, err := app.db.GetUserByEmail(form.Email)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		form.Validator.CheckField(form.Email != "", "Email", "Email is required")
		form.Validator.CheckField(validator.Matches(form.Email, validator.RgxEmail), "Email", "Must be a valid email address")
		form.Validator.CheckField(found, "Email", "No matching email found")

		if form.Validator.HasErrors() {
			data := app.newTemplateData(r)
			data["Form"] = form

			err := response.Page(w, http.StatusUnprocessableEntity, data, "pages/forgotten-password.tmpl")
			if err != nil {
				app.serverError(w, r, err)
			}
			return
		}

		plaintextToken := token.New()

		hashedToken := token.Hash(plaintextToken)

		err = app.db.InsertPasswordReset(hashedToken, user.ID, 24*time.Hour)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		data := app.newEmailData()
		data["PlaintextToken"] = plaintextToken

		err = app.mailer.Send(user.Email, data, "forgotten-password.tmpl")
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		http.Redirect(w, r, "/forgotten-password-confirmation", http.StatusSeeOther)
	}
}

func (app *application) forgottenPasswordConfirmation(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	err := response.Page(w, http.StatusOK, data, "pages/forgotten-password-confirmation.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) passwordReset(w http.ResponseWriter, r *http.Request) {
	plaintextToken := r.PathValue("plaintextToken")

	hashedToken := token.Hash(plaintextToken)

	passwordReset, found, err := app.db.GetPasswordReset(hashedToken)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if !found {
		data := app.newTemplateData(r)
		data["InvalidLink"] = true

		err := response.Page(w, http.StatusUnprocessableEntity, data, "pages/password-reset.tmpl")
		if err != nil {
			app.serverError(w, r, err)
		}
		return
	}

	var form struct {
		NewPassword string              `form:"NewPassword"`
		Validator   validator.Validator `form:"-"`
	}

	switch r.Method {
	case http.MethodGet:
		data := app.newTemplateData(r)
		data["Form"] = form
		data["PlaintextToken"] = plaintextToken

		err := response.Page(w, http.StatusOK, data, "pages/password-reset.tmpl")
		if err != nil {
			app.serverError(w, r, err)
		}

	case http.MethodPost:
		err := request.DecodePostForm(r, &form)
		if err != nil {
			app.badRequest(w, r, err)
			return
		}

		form.Validator.CheckField(form.NewPassword != "", "NewPassword", "La contraseña es obligatoria")
		form.Validator.CheckField(len(form.NewPassword) >= 8, "NewPassword", "La contraseña debe tener al menos 8 caracteres")
		form.Validator.CheckField(len(form.NewPassword) <= 72, "NewPassword", "La contraseña es demasiado larga (máximo 72 caracteres)")
		form.Validator.CheckField(validator.NotIn(form.NewPassword, password.CommonPasswords...), "NewPassword", "Esta contraseña es muy común. Usa una más segura")

		if form.Validator.HasErrors() {
			data := app.newTemplateData(r)
			data["Form"] = form
			data["PlaintextToken"] = plaintextToken

			err := response.Page(w, http.StatusUnprocessableEntity, data, "pages/password-reset.tmpl")
			if err != nil {
				app.serverError(w, r, err)
			}
			return
		}

		hashedPassword, err := password.Hash(form.NewPassword)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		err = app.db.UpdateUserHashedPassword(passwordReset.UserID, hashedPassword)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		err = app.db.DeletePasswordResets(passwordReset.UserID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		http.Redirect(w, r, "/password-reset-confirmation", http.StatusSeeOther)
	}
}

func (app *application) passwordResetConfirmation(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	err := response.Page(w, http.StatusOK, data, "pages/password-reset-confirmation.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) restricted(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	err := response.Page(w, http.StatusOK, data, "pages/restricted.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) salaryCalculator(w http.ResponseWriter, r *http.Request) {
	var form struct {
		GrossMonthlySalary float64             `form:"GrossMonthlySalary"`
		YearsOfService     int                 `form:"YearsOfService"`
		Validator          validator.Validator `form:"-"`
	}

	switch r.Method {
	case http.MethodGet:
		data := app.newTemplateData(r)
		data["Form"] = form

		err := response.Page(w, http.StatusOK, data, "pages/calculator.tmpl")
		if err != nil {
			app.serverError(w, r, err)
		}

	case http.MethodPost:
		err := request.DecodePostForm(r, &form)
		if err != nil {
			app.badRequest(w, r, err)
			return
		}

		form.Validator.CheckField(form.GrossMonthlySalary > 0, "GrossMonthlySalary", "El salario debe ser mayor a 0")
		form.Validator.CheckField(form.GrossMonthlySalary <= 1000000, "GrossMonthlySalary", "El salario es demasiado alto")
		form.Validator.CheckField(form.YearsOfService >= 0, "YearsOfService", "Los años de servicio no pueden ser negativos")

		if form.Validator.HasErrors() {
			data := app.newTemplateData(r)
			data["Form"] = form

			err := response.Page(w, http.StatusUnprocessableEntity, data, "pages/calculator.tmpl")
			if err != nil {
				app.serverError(w, r, err)
			}
			return
		}

		// Get fiscal year configuration
		fiscalYear, found, err := app.db.GetActiveFiscalYear()
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		if !found {
			app.serverError(w, r, err)
			return
		}

		// Calculate salary
		result, err := app.calculateSalary(form.GrossMonthlySalary, form.YearsOfService, fiscalYear)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		data := app.newTemplateData(r)
		data["Form"] = form
		data["Result"] = result
		data["FiscalYear"] = fiscalYear

		err = response.Page(w, http.StatusOK, data, "pages/calculator.tmpl")
		if err != nil {
			app.serverError(w, r, err)
		}
	}
}

// privacy displays the privacy policy (Aviso de Privacidad)
func (app *application) privacy(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	
	err := response.Page(w, http.StatusOK, data, "pages/privacy.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

// terms displays the terms and conditions (Términos y Condiciones)
func (app *application) terms(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	
	err := response.Page(w, http.StatusOK, data, "pages/terms.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

// robotsTxt serves the robots.txt file for SEO
func (app *application) robotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	
	robotsContent := `User-agent: *
Allow: /

Sitemap: https://totalcomp.mx/sitemap.xml`
	
	w.Write([]byte(robotsContent))
}

// sitemapXML serves the sitemap.xml file for SEO
func (app *application) sitemapXML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	
	sitemapContent := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://totalcomp.mx/</loc>
    <lastmod>2025-11-19</lastmod>
    <changefreq>weekly</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>https://totalcomp.mx/privacy</loc>
    <lastmod>2025-11-19</lastmod>
    <changefreq>monthly</changefreq>
    <priority>0.5</priority>
  </url>
  <url>
    <loc>https://totalcomp.mx/terms</loc>
    <lastmod>2025-11-19</lastmod>
    <changefreq>monthly</changefreq>
    <priority>0.5</priority>
  </url>
</urlset>`
	
	w.Write([]byte(sitemapContent))
}

// accountDeveloper displays the developer dashboard where users can manage their API keys
func (app *application) accountDeveloper(w http.ResponseWriter, r *http.Request) {
	authenticatedUser, found := contextGetAuthenticatedUser(r)
	if !found {
		app.notFound(w, r)
		return
	}
	userID := authenticatedUser.ID
	
	user, found, err := app.db.GetUser(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if !found {
		app.notFound(w, r)
		return
	}
	
	data := app.newTemplateData(r)
	data["User"] = user
	
	err = response.Page(w, http.StatusOK, data, "pages/developer.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

// generateAPIKey generates or regenerates an API key for the authenticated user
func (app *application) generateAPIKey(w http.ResponseWriter, r *http.Request) {
	authenticatedUser, found := contextGetAuthenticatedUser(r)
	if !found {
		app.notFound(w, r)
		return
	}
	userID := authenticatedUser.ID
	
	// Generate a secure random API key (32 characters)
	apiKey, err := app.generateSecureAPIKey()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	
	// Store the API key in the database (plain text for now, could hash later)
	err = app.db.UpdateUserAPIKey(userID, apiKey)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	
	// Redirect back to developer dashboard
	http.Redirect(w, r, "/account/developer", http.StatusSeeOther)
}

// developersPage renders the public marketing page for the API
func (app *application) developersPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	
	err := response.Page(w, http.StatusOK, data, "pages/developers.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

// verifyEmail handles the email verification flow
func (app *application) verifyEmail(w http.ResponseWriter, r *http.Request) {
	// Get verification token from URL
	plaintextToken := r.PathValue("plaintextToken")

	// Hash the token to compare with database
	hashedToken := token.Hash(plaintextToken)

	// Get user ID from token (validates expiry - 24 hours)
	userID, found, err := app.db.GetUserIDFromVerificationToken(hashedToken)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if !found {
		data := app.newTemplateData(r)
		data["Message"] = "El enlace de verificación es inválido o ha expirado."
		err := response.Page(w, http.StatusBadRequest, data, "pages/email-verification-error.tmpl")
		if err != nil {
			app.serverError(w, r, err)
		}
		return
	}

	// Mark email as verified
	err = app.db.VerifyUserEmail(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Show success page
	data := app.newTemplateData(r)
	err = response.Page(w, http.StatusOK, data, "pages/email-verification-success.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

// resendVerificationEmail generates a new verification token and sends it to the user's email
func (app *application) resendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from session
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	
	// Get user email from database
	user, found, err := app.db.GetUser(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	
	if !found {
		app.sessionManager.Put(r.Context(), "flash", "Usuario no encontrado")
		http.Redirect(w, r, "/account/developer", http.StatusSeeOther)
		return
	}
	
	// Check if already verified
	if user.EmailVerified {
		app.sessionManager.Put(r.Context(), "flash", "Tu email ya está verificado")
		http.Redirect(w, r, "/account/developer", http.StatusSeeOther)
		return
	}
	
	// Generate new verification token
	plaintextToken := token.New()
	hashedToken := token.Hash(plaintextToken)
	
	// Delete any existing verification tokens for this user
	err = app.db.DeleteEmailVerificationTokensForUser(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	
	// Store new verification token in database
	err = app.db.InsertEmailVerificationToken(userID, hashedToken)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	
	// Send verification email in background
	app.backgroundTask(r, func() error {
		data := app.newEmailData()
		data["Email"] = user.Email
		data["VerificationToken"] = plaintextToken
		return app.mailer.Send(user.Email, data, "welcome.tmpl")
	})
	
	app.sessionManager.Put(r.Context(), "flash", "Email de verificación reenviado. Revisa tu bandeja de entrada.")
	http.Redirect(w, r, "/account/developer", http.StatusSeeOther)
}

// apiCalculate is the main API endpoint for salary calculations (JSON API)
func (app *application) apiCalculate(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request body
	var req struct {
		Salary                  float64 `json:"salary"`
		Regime                  string  `json:"regime"` // "sueldos" or "resico"
		HasAguinaldo            bool    `json:"has_aguinaldo"`
		AguinaldoDays           int     `json:"aguinaldo_days"`
		HasValesDespensa        bool    `json:"has_vales_despensa"`
		ValesDespensaAmount     float64 `json:"vales_despensa_amount"`
		HasPrimaVacacional      bool    `json:"has_prima_vacacional"`
		VacationDays            int     `json:"vacation_days"`
		PrimaVacacionalPercent  float64 `json:"prima_vacacional_percent"`
		HasFondoAhorro          bool    `json:"has_fondo_ahorro"`
		FondoAhorroPercent      float64 `json:"fondo_ahorro_percent"`
		UnpaidVacationDays      int     `json:"unpaid_vacation_days"` // RESICO only
	}
	
	err := request.DecodeJSON(w, r, &req)
	if err != nil {
		err := response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON request body",
		})
		if err != nil {
			app.serverError(w, r, err)
		}
		return
	}
	
	// Validate input
	if req.Salary <= 0 {
		err := response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "Salary must be greater than 0",
		})
		if err != nil {
			app.serverError(w, r, err)
		}
		return
	}
	
	// Get fiscal year configuration
	fiscalYear, found, err := app.db.GetActiveFiscalYear()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if !found {
		err := response.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": "No active fiscal year configuration found",
		})
		if err != nil {
			app.serverError(w, r, err)
		}
		return
	}
	
	// Calculate based on regime
	var result database.SalaryCalculation
	
	if req.Regime == "resico" {
		// RESICO calculation
		result, err = app.calculateRESICO(req.Salary, req.UnpaidVacationDays, []OtherBenefit{}, 1.0, fiscalYear)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	} else {
		// Sueldos y Salarios calculation (default)
		result, err = app.calculateSalaryWithBenefits(
			req.Salary,
			req.HasAguinaldo, req.AguinaldoDays,
			req.HasValesDespensa, req.ValesDespensaAmount,
			req.HasPrimaVacacional, req.VacationDays, req.PrimaVacacionalPercent,
			req.HasFondoAhorro, req.FondoAhorroPercent,
			false, // hasInfonavitCredit - API users can add this later
			[]OtherBenefit{},
			1.0, // Exchange rate (MXN)
			fiscalYear,
		)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}
	
	// Return JSON response
	jsonResponse := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"regime":            req.Regime,
			"gross_salary":      result.GrossSalary,
			"net_salary":        result.NetSalary,
			"isr_tax":           result.ISRTax,
			"subsidio_empleo":   result.SubsidioEmpleo,
			"imss_worker":       result.IMSSWorker,
			"sbc":               result.SBC,
			"yearly_gross_base": result.YearlyGrossBase,
			"yearly_gross":      result.YearlyGross,
			"yearly_net":        result.YearlyNet,
			"monthly_adjusted":  result.MonthlyAdjusted,
			"breakdown": map[string]interface{}{
				"aguinaldo_gross": result.AguinaldoGross,
				"aguinaldo_isr":   result.AguinaldoISR,
				"aguinaldo_net":   result.AguinaldoNet,
				"prima_vacacional_gross": result.PrimaVacacionalGross,
				"prima_vacacional_isr":   result.PrimaVacacionalISR,
				"prima_vacacional_net":   result.PrimaVacacionalNet,
				"fondo_ahorro_yearly":    result.FondoAhorroYearly,
				"infonavit_employer_annual": result.InfonavitEmployerAnnual,
				"imss_employer_annual":      result.IMSSEmployerAnnual,
			},
		},
		"meta": map[string]interface{}{
			"fiscal_year": fiscalYear.Year,
			"uma_monthly": fiscalYear.UMAMonthly,
		},
	}
	
	err = response.JSON(w, http.StatusOK, jsonResponse)
	if err != nil {
		app.serverError(w, r, err)
	}
}

// exportPDF generates and downloads a comparison PDF report for all packages
func (app *application) exportPDF(w http.ResponseWriter, r *http.Request) {
	// Get results from session
	if !app.sessionManager.Exists(r.Context(), "comparisonResults") {
		app.badRequest(w, r, fmt.Errorf("no results in session"))
		return
	}
	
	results, ok := app.sessionManager.Get(r.Context(), "comparisonResults").([]PackageResult)
	if !ok {
		app.serverError(w, r, fmt.Errorf("invalid results in session"))
		return
	}
	
	if len(results) == 0 {
		app.badRequest(w, r, fmt.Errorf("no packages to compare"))
		return
	}
	
	// Get package inputs from session
	packageInputs, ok := app.sessionManager.Get(r.Context(), "packageInputs").([]PackageInput)
	if !ok {
		app.serverError(w, r, fmt.Errorf("invalid package inputs in session"))
		return
	}
	
	// Get fiscal year from session
	fiscalYear, ok := app.sessionManager.Get(r.Context(), "fiscalYear").(database.FiscalYear)
	if !ok {
		// Try to get from database if not in session
		fy, found, err := app.db.GetActiveFiscalYear()
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		if !found {
			app.serverError(w, r, fmt.Errorf("no active fiscal year"))
			return
		}
		fiscalYear = fy
	}
	
	// Convert results to pdf.PackageResult format (merge inputs + calculations)
	pdfPackages := make([]pdf.PackageResult, len(results))
	for i, result := range results {
		// Convert OtherBenefit to pdf.OtherBenefit
		var pdfOtherBenefits []pdf.OtherBenefit
		if i < len(packageInputs) {
			for _, ob := range packageInputs[i].OtherBenefits {
				pdfOtherBenefits = append(pdfOtherBenefits, pdf.OtherBenefit{
					Name:     ob.Name,
					Amount:   ob.Amount,
					TaxFree:  ob.TaxFree,
					Currency: ob.Currency,
					Cadence:  ob.Cadence,
				})
			}
		}
		
		pdfInput := pdf.PackageInput{}
		if i < len(packageInputs) {
			pdfInput = pdf.PackageInput{
				Name:                    packageInputs[i].Name,
				Regime:                  packageInputs[i].Regime,
				Currency:                packageInputs[i].Currency,
				ExchangeRate:            packageInputs[i].ExchangeRate,
				PaymentFrequency:        packageInputs[i].PaymentFrequency,
				HoursPerWeek:            packageInputs[i].HoursPerWeek,
				GrossMonthlySalary:      packageInputs[i].GrossMonthlySalary,
				HasAguinaldo:            packageInputs[i].HasAguinaldo,
				AguinaldoDays:           packageInputs[i].AguinaldoDays,
				HasValesDespensa:        packageInputs[i].HasValesDespensa,
				ValesDespensaAmount:     packageInputs[i].ValesDespensaAmount,
				HasPrimaVacacional:      packageInputs[i].HasPrimaVacacional,
				VacationDays:            packageInputs[i].VacationDays,
				PrimaVacacionalPercent:  packageInputs[i].PrimaVacacionalPercent,
				HasFondoAhorro:          packageInputs[i].HasFondoAhorro,
				FondoAhorroPercent:      packageInputs[i].FondoAhorroPercent,
				UnpaidVacationDays:      packageInputs[i].UnpaidVacationDays,
				OtherBenefits:           pdfOtherBenefits,
				HasEquity:               packageInputs[i].HasEquity,
				InitialEquityUSD:        packageInputs[i].InitialEquityUSD,
				HasRefreshers:           packageInputs[i].HasRefreshers,
				RefresherMinUSD:         packageInputs[i].RefresherMinUSD,
				RefresherMaxUSD:         packageInputs[i].RefresherMaxUSD,
			}
		}
		
		pdfPackages[i] = pdf.PackageResult{
			Name:        result.PackageName,
			Input:       pdfInput,
			Calculation: result.SalaryCalculation,
		}
	}
	
	// Generate comparison PDF
	pdfBytes, err := pdf.GenerateComparisonReport(pdfPackages, fiscalYear)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	
	// Set headers for PDF download
	w.Header().Set("Content-Type", "application/pdf")
	filename := "TotalComp_Comparacion_2025.pdf"
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	
	// Write PDF bytes to response
	_, err = w.Write(pdfBytes)
	if err != nil {
		app.logger.Error("failed to write PDF response", "error", err)
	}
}

// sanitizeFilename removes special characters from filename
func sanitizeFilename(name string) string {
	if name == "" {
		return "Paquete"
	}
	// Simple sanitization - replace spaces and special chars
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result += string(r)
		} else if r == ' ' {
			result += "_"
		}
	}
	if result == "" {
		return "Paquete"
	}
	return result
}
