package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"time"

	"local.io/go-astro-re/internal/domain"
)

type Service interface {
	Evaluate(context.Context, domain.AstrologyInput, int) (domain.EvaluationReport, error)
}

type Runner struct {
	service Service
	out     io.Writer
}

func NewRunner(service Service, out io.Writer) Runner {
	return Runner{service: service, out: out}
}

func (r Runner) Run(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("vedic-eval", flag.ContinueOnError)
	datetime := fs.String("datetime", "", "RFC3339 datetime")
	timezone := fs.String("timezone", "Asia/Kolkata", "timezone")
	locationName := fs.String("location-name", "Delhi", "location name")
	countryCode := fs.String("country-code", "IN", "country code")
	latitude := fs.Float64("lat", 28.6139, "latitude")
	longitude := fs.Float64("lon", 77.2090, "longitude")
	profile := fs.String("profile", "default", "calculation profile")
	workers := fs.Int("workers", 4, "worker count")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *datetime == "" {
		return fmt.Errorf("datetime is required")
	}
	dt, err := time.Parse(time.RFC3339, *datetime)
	if err != nil {
		return fmt.Errorf("parse datetime: %w", err)
	}

	report, err := r.service.Evaluate(ctx, domain.AstrologyInput{
		DateTime: dt,
		Timezone: *timezone,
		Location: domain.Location{
			Name:        *locationName,
			Latitude:    *latitude,
			Longitude:   *longitude,
			CountryCode: *countryCode,
		},
		CalculationProfile: *profile,
	}, *workers)
	if err != nil {
		return err
	}

	return json.NewEncoder(r.out).Encode(report)
}
