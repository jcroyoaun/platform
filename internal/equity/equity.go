package equity

import "math"

// YearlyEquity represents equity vesting for a single year
type YearlyEquity struct {
	Year                int
	InitialGrantVested  float64         // Amount vested from initial grant (USD)
	RefresherVested     map[int]float64 // key = refresher year (1,2,3...), value = amount vested (USD)
	RefresherTotal      float64         // Total from all refreshers this year (USD)
	TotalVested         float64         // Total vested this year (USD)
	NewRefresherGranted float64         // Refresher granted THIS year (USD, starts vesting next year)
	TotalVestedMXN      float64         // Total vested in MXN using exchange rate
}

// EquityConfig holds the configuration for equity calculations
type EquityConfig struct {
	InitialGrantUSD float64
	HasRefreshers   bool
	RefresherMinUSD float64
	RefresherMaxUSD float64
	VestingYears    int     // Typically 4
	ExchangeRate    float64 // From FiscalYear.USDMXNRate
}

// CalculateEquitySchedule calculates year-by-year equity vesting with refresher stacking
// Returns a slice of YearlyEquity for the specified number of years (including Year 0)
func CalculateEquitySchedule(config EquityConfig, years int) []YearlyEquity {
	schedule := make([]YearlyEquity, years+1) // +1 for Year 0
	
	// Calculate average refresher amount
	avgRefresher := 0.0
	if config.HasRefreshers {
		avgRefresher = (config.RefresherMinUSD + config.RefresherMaxUSD) / 2.0
	}
	
	// Annual vesting percentage (typically 25% per year for 4 years)
	annualVestPercent := 1.0 / float64(config.VestingYears)
	
	// Year 0: Join date - grant awarded but nothing vests yet
	schedule[0] = YearlyEquity{
		Year:                0,
		InitialGrantVested:  0,
		RefresherVested:     make(map[int]float64),
		RefresherTotal:      0,
		TotalVested:         0,
		NewRefresherGranted: 0,
		TotalVestedMXN:      0,
	}
	
	// Track refresher grants by year they were granted
	// refresherGrants[grantYear] = amount
	refresherGrants := make(map[int]float64)
	
	for year := 1; year <= years; year++ {
		yearEquity := YearlyEquity{
			Year:            year,
			RefresherVested: make(map[int]float64),
		}
		
		// 1. Vest from initial grant (only for first VestingYears years)
		if year <= config.VestingYears {
			yearEquity.InitialGrantVested = config.InitialGrantUSD * annualVestPercent
		}
		
		// 2. Vest from refreshers (they vest starting 1 year after granted)
		if config.HasRefreshers {
			for grantYear, grantAmount := range refresherGrants {
				vestingStartYear := grantYear + 1
				vestingEndYear := grantYear + config.VestingYears
				
				// Check if this grant is currently vesting
				if year >= vestingStartYear && year <= vestingEndYear {
					vestedAmount := grantAmount * annualVestPercent
					yearEquity.RefresherVested[grantYear] = vestedAmount
				}
			}
		}
		
		// 3. Grant new refresher this year (starts vesting next year)
		if config.HasRefreshers {
			yearEquity.NewRefresherGranted = avgRefresher
			refresherGrants[year] = avgRefresher
		}
		
		// 4. Calculate total from refreshers
		yearEquity.RefresherTotal = 0.0
		for _, amount := range yearEquity.RefresherVested {
			yearEquity.RefresherTotal += amount
		}
		
		// 5. Calculate total vested this year (USD)
		yearEquity.TotalVested = yearEquity.InitialGrantVested + yearEquity.RefresherTotal
		
		// 6. Convert to MXN
		yearEquity.TotalVestedMXN = yearEquity.TotalVested * config.ExchangeRate
		
		// Round to 2 decimal places
		yearEquity.TotalVested = math.Round(yearEquity.TotalVested*100) / 100
		yearEquity.TotalVestedMXN = math.Round(yearEquity.TotalVestedMXN*100) / 100
		yearEquity.InitialGrantVested = math.Round(yearEquity.InitialGrantVested*100) / 100
		yearEquity.RefresherTotal = math.Round(yearEquity.RefresherTotal*100) / 100
		yearEquity.NewRefresherGranted = math.Round(yearEquity.NewRefresherGranted*100) / 100
		
		for k, v := range yearEquity.RefresherVested {
			yearEquity.RefresherVested[k] = math.Round(v*100) / 100
		}
		
		schedule[year] = yearEquity
	}
	
	return schedule
}

// GetTotalEquityOver4Years returns the total equity vested over 4 years
func GetTotalEquityOver4Years(config EquityConfig) (float64, float64) {
	schedule := CalculateEquitySchedule(config, 4)
	
	totalUSD := 0.0
	totalMXN := 0.0
	
	for _, year := range schedule {
		totalUSD += year.TotalVested
		totalMXN += year.TotalVestedMXN
	}
	
	return totalUSD, totalMXN
}

