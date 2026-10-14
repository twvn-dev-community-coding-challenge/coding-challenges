package phone

import (
	"errors"
	"strings"
	"unicode"
)

var (
	ErrInvalidCountry = errors.New("unsupported or invalid country")
	ErrInvalidNumber  = errors.New("invalid phone number format")
)

// ISO 3166-1 alpha-2 codes supported by the SMS router.
var dialCodeByISO2 = map[string]string{
	"VN": "84",
	"TH": "66",
	"SG": "65",
	"PH": "63",
}

var legacyCountryNames = map[string]string{
	"việt nam":  "VN",
	"vietnam":   "VN",
	"thailand":  "TH",
	"singapore": "SG",
	"philippines": "PH",
	"philippine":  "PH",
}

// NormalizeCountry maps user input to ISO2 (e.g. PH, ph, Philippines -> PH).
func NormalizeCountry(country string) (string, error) {
	s := strings.TrimSpace(country)
	if s == "" {
		return "", ErrInvalidCountry
	}
	if len(s) == 2 {
		u := strings.ToUpper(s)
		if _, ok := dialCodeByISO2[u]; ok {
			return u, nil
		}
		return "", ErrInvalidCountry
	}
	key := strings.ToLower(s)
	if iso, ok := legacyCountryNames[key]; ok {
		return iso, nil
	}
	return "", ErrInvalidCountry
}

// CanonicalDigits returns E.164 digits only (no '+').
func CanonicalDigits(phone string) string {
	var b strings.Builder
	for _, r := range phone {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Validate checks country and phone, returns canonical digits only (E.164 without '+').
// Accepts:
//   - Full international form already starting with the country calling code (e.g. PH: 63917xxxxxxx).
//   - National form with trunk prefix 0 (e.g. PH: 09171234567 -> 639171234567).
//   - National digits without country code or 0 (e.g. PH: 9171234567 -> 639171234567).
// If digits after a leading 0 already begin with the CC (e.g. 0639171234567), that remainder is used.
func Validate(countryISO2, phone string) (canonical string, err error) {
	cc, ok := dialCodeByISO2[countryISO2]
	if !ok {
		return "", ErrInvalidCountry
	}
	digits := CanonicalDigits(phone)
	if len(digits) == 0 {
		return "", ErrInvalidNumber
	}

	if strings.HasPrefix(digits, cc) {
		return finalizeCanonical(cc, digits)
	}

	// National dialing: leading 0 (e.g. 09xx… PH, 08xx… TH), then prepend country code.
	if strings.HasPrefix(digits, "0") {
		national := digits[1:]
		if national == "" {
			return "", ErrInvalidNumber
		}
		if strings.HasPrefix(national, cc) {
			digits = national
		} else {
			digits = cc + national
		}
		return finalizeCanonical(cc, digits)
	}

	// National / arbitrary digits: assume selected country and prepend its calling code.
	digits = cc + digits
	return finalizeCanonical(cc, digits)
}

func finalizeCanonical(cc, digits string) (string, error) {
	if !strings.HasPrefix(digits, cc) {
		digits = cc + digits
	}
	if len(digits) < 8 || len(digits) > 15 {
		return "", ErrInvalidNumber
	}
	return digits, nil
}
