package pdf

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"github.com/jcroyoaun/totalcompmx/internal/database"
)

// OtherBenefit represents custom benefits
type OtherBenefit struct {
	Name     string
	Amount   float64
	TaxFree  bool
	Currency string
	Cadence  string
}

// PackageInput represents the input details for a package
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
	UnpaidVacationDays      string
	OtherBenefits           []OtherBenefit
	// Equity fields
	HasEquity               bool
	InitialEquityUSD        string
	HasRefreshers           bool
	RefresherMinUSD         string
	RefresherMaxUSD         string
}

// PackageResult represents a single package's calculation results
type PackageResult struct {
	Name        string
	Input       PackageInput
	Calculation *database.SalaryCalculation
}

// ReportData represents the data passed to the PDF template
type ReportData struct {
	Date     string
	Packages []PackageResult
}

// GenerateComparisonReport generates a PDF by rendering an HTML template via Chrome
func GenerateComparisonReport(packages []PackageResult, fiscalYear database.FiscalYear) ([]byte, error) {
	// Prepare template data
	data := ReportData{
		Date:     time.Now().Format("02 Jan 2006"),
		Packages: packages,
	}

	// Custom template functions
	funcMap := template.FuncMap{
		"formatFloat": func(f float64, decimals int) string {
			// Format with thousands separator
			formatted := fmt.Sprintf(fmt.Sprintf("%%.%df", decimals), f)
			return addThousandsSeparator(formatted)
		},
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"mul": func(a, b float64) float64 {
			return a * b
		},
	}

	// Find template path (works both in dev and production)
	templatePath := "assets/templates/pdf/report.tmpl"
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// Try from working directory
		wd, _ := os.Getwd()
		templatePath = filepath.Join(wd, templatePath)
	}

	// Parse the template
	tmpl, err := template.New("report.tmpl").Funcs(funcMap).ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PDF template at %s: %w", templatePath, err)
	}

	// Render the template to HTML string
	var htmlBuf bytes.Buffer
	if err := tmpl.Execute(&htmlBuf, data); err != nil {
		return nil, fmt.Errorf("failed to execute PDF template: %w", err)
	}

	htmlContent := htmlBuf.String()

	// Generate PDF using Chrome
	pdfBytes, err := renderHTMLToPDF(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to render HTML to PDF: %w", err)
	}

	return pdfBytes, nil
}

// renderHTMLToPDF uses chromedp to render HTML to PDF
func renderHTMLToPDF(htmlContent string) ([]byte, error) {
	// Create Chrome context with options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Headless,
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var pdfBuf []byte

	// Navigate to about:blank and set HTML content, then print to PDF
	err := chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Set the document content
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}

			return page.SetDocumentContent(frameTree.Frame.ID, htmlContent).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Wait a bit for rendering
			time.Sleep(500 * time.Millisecond)

			// Print to PDF with options
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithPreferCSSPageSize(true).
				WithScale(0.75).
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				Do(ctx)
			return err
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("chromedp error: %w", err)
	}

	return pdfBuf, nil
}

// addThousandsSeparator adds commas to number strings for readability
func addThousandsSeparator(s string) string {
	// Split by decimal point
	parts := strings.Split(s, ".")
	intPart := parts[0]
	decPart := ""
	if len(parts) > 1 {
		decPart = "." + parts[1]
	}

	// Handle negative numbers
	negative := false
	if strings.HasPrefix(intPart, "-") {
		negative = true
		intPart = intPart[1:]
	}

	// Add commas every 3 digits from the right
	var result []rune
	for i, char := range reverse(intPart) {
		if i > 0 && i%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, char)
	}

	finalInt := reverse(string(result))
	if negative {
		finalInt = "-" + finalInt
	}

	return finalInt + decPart
}

// reverse reverses a string
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
