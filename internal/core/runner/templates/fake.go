package templates

import (
	"math"
	"strconv"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
)

func fakeName(_ ...string) (any, error) {
	return gofakeit.Name(), nil
}

func fakeEmail(_ ...string) (any, error) {
	return gofakeit.Email(), nil
}

func fakePassword(args ...string) (any, error) {
	length := 12
	lower := true
	upper := true
	numeric := true
	special := false
	space := false

	for _, arg := range args {
		arg = strings.ToLower(strings.TrimSpace(arg))

		switch arg {
		case "nolower":
			lower = false
		case "noupper":
			upper = false
		case "nonumeric":
			numeric = false
		case "special":
			special = true
		case "space":
			space = true
		default:
			if n, err := strconv.Atoi(arg); err == nil {
				length = n
			}
		}
	}

	return gofakeit.Password(lower, upper, numeric, special, space, length), nil
}

func fakeInt(args ...string) (any, error) {
	minInt, maxInt := -1_000_000_000_000_000, 1_000_000_000_000_000

	if len(args) > 0 {
		n, err := strconv.ParseInt(args[0], 10, 64)
		if err == nil {
			minInt = int(n)
		}

		if len(args) > 1 {
			n, err = strconv.ParseInt(args[1], 10, 64)
			if err == nil {
				maxInt = int(n)
			}
		}
	}

	if minInt > maxInt {
		minInt, maxInt = maxInt, minInt-1
	}

	return gofakeit.IntRange(minInt, maxInt), nil
}

func fakeUint(args ...string) (any, error) {
	var minInt, maxInt uint64 = 0, math.MaxUint

	if len(args) > 0 {
		n, err := strconv.ParseUint(args[0], 10, 64)
		if err == nil {
			minInt = n
		}

		if len(args) > 1 {
			n, err = strconv.ParseUint(args[1], 10, 64)
			if err == nil {
				maxInt = n
			}
		}
	}

	if minInt >= maxInt {
		minInt, maxInt = maxInt, minInt
	}

	return gofakeit.UintRange(uint(minInt), uint(maxInt)), nil
}

func fakeFloat(_ ...string) (any, error) {
	return gofakeit.Float64(), nil
}

func fakeBool(_ ...string) (any, error) {
	return gofakeit.Bool(), nil
}

func fakePhone(_ ...string) (any, error) {
	return gofakeit.Phone(), nil
}

func fakeAddress(_ ...string) (any, error) {
	return gofakeit.Address().Address, nil
}

func fakeCompany(_ ...string) (any, error) {
	return gofakeit.Company(), nil
}

func fakeDate(_ ...string) (any, error) {
	return gofakeit.Date(), nil
}

func fakeUUID(_ ...string) (any, error) {
	return gofakeit.UUID(), nil
}

func fakeURL(_ ...string) (any, error) {
	return gofakeit.URL(), nil
}

func fakeColor(_ ...string) (any, error) {
	return gofakeit.Color(), nil
}

func fakeWord(_ ...string) (any, error) {
	return gofakeit.Word(), nil
}

func fakeSentence(_ ...string) (any, error) {
	return gofakeit.SentenceSimple(), nil
}

func fakeCountry(_ ...string) (any, error) {
	return gofakeit.Country(), nil
}

func fakeCity(_ ...string) (any, error) {
	return gofakeit.City(), nil
}
