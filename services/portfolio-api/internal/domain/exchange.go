package domain

// Exchange is reference data for a trading venue identified by its ISO 10383 MIC code.
type Exchange struct {
	Code     string // ISO 10383 MIC, e.g. "XNAS"
	Name     string // e.g. "NASDAQ Stock Market"
	Country  string // ISO 3166-1 alpha-2, e.g. "US"
	Timezone string // IANA timezone, e.g. "America/New_York"
}
