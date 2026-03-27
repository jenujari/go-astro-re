package mock

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"local.io/go-astro-re/internal/domain"
)

type Builder struct{}

func NewBuilder() Builder {
	return Builder{}
}

func (Builder) Build(_ context.Context, input domain.AstrologyInput) (domain.AstrologyContext, error) {
	seed := stableSeed(input)

	moonNakshatra := choose(seed, []string{"Rohini", "Ashwini", "Hasta", "Swati", "Pushya"})
	saturnAspectsMoon := seed%2 == 0
	marsOwnSign := seed%3 == 0
	jupiterExalted := seed%5 == 0
	sunDebilitated := seed%7 == 0

	positions := []domain.PlanetPosition{
		{Planet: "Moon", Sign: choose(seed+1, []string{"Taurus", "Cancer", "Libra"}), House: int((seed % 12) + 1), Nakshatra: moonNakshatra},
		{Planet: "Saturn", Sign: choose(seed+2, []string{"Aquarius", "Capricorn", "Pisces"}), House: int(((seed + 3) % 12) + 1), Retrograde: seed%4 == 0},
		{Planet: "Mars", Sign: choose(seed+3, []string{"Aries", "Scorpio", "Gemini"}), House: int(((seed + 5) % 12) + 1), IsOwnSign: marsOwnSign},
		{Planet: "Jupiter", Sign: choose(seed+4, []string{"Cancer", "Sagittarius", "Pisces"}), House: int(((seed + 7) % 12) + 1), IsExalted: jupiterExalted},
		{Planet: "Sun", Sign: choose(seed+5, []string{"Libra", "Leo", "Aries"}), House: int(((seed + 9) % 12) + 1), IsDebilitated: sunDebilitated},
	}

	facts := map[string]domain.DerivedFact{
		"moon_nakshatra": {
			Key: "moon_nakshatra", Value: moonNakshatra, Source: "mock-builder",
		},
		"saturn_aspects_moon": {
			Key: "saturn_aspects_moon", BoolValue: boolPtr(saturnAspectsMoon), Source: "mock-builder",
		},
		"mars_own_sign": {
			Key: "mars_own_sign", BoolValue: boolPtr(marsOwnSign), Source: "mock-builder",
		},
		"jupiter_exalted": {
			Key: "jupiter_exalted", BoolValue: boolPtr(jupiterExalted), Source: "mock-builder",
		},
		"sun_debilitated": {
			Key: "sun_debilitated", BoolValue: boolPtr(sunDebilitated), Source: "mock-builder",
		},
		"context_seed": {
			Key: "context_seed", Value: fmt.Sprintf("%d", seed), Source: "mock-builder",
		},
	}

	scoreStrength := float64((seed % 10) + 1)
	facts["benefic_strength"] = domain.DerivedFact{
		Key: "benefic_strength", NumericValue: &scoreStrength, Source: "mock-builder",
	}

	return domain.AstrologyContext{
		Input:           input,
		PlanetPositions: positions,
		DerivedFacts:    domain.DerivedFacts{Items: facts},
		BuilderVersion:  "mock-vedic-v1",
	}, nil
}

func stableSeed(input domain.AstrologyInput) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(input.DateTime.In(time.UTC).Format(time.RFC3339)))
	_, _ = h.Write([]byte(input.Timezone))
	_, _ = h.Write([]byte(input.Location.Name))
	_, _ = h.Write([]byte(input.Location.CountryCode))
	_, _ = h.Write([]byte(input.CalculationProfile))
	return h.Sum32()
}

func choose[T any](seed uint32, items []T) T {
	return items[int(seed)%len(items)]
}

func boolPtr(v bool) *bool {
	return &v
}
