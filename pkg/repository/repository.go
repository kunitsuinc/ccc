package repository

import (
	"context"
	"time"

	"github.com/kunitsuinc/ccc/pkg/domain"
	"github.com/kunitsuinc/ccc/pkg/errorz"
	"github.com/kunitsuinc/ccc/pkg/repository/bigquery"
	"github.com/kunitsuinc/util.go/slice"
)

type Repository struct {
	bigquery *bigquery.BigQuery
}

type Option func(r *Repository) *Repository

func New(opts ...Option) *Repository {
	r := &Repository{}

	for _, opt := range opts {
		r = opt(r)
	}

	return r
}

func WithBigQuery(bigquery *bigquery.BigQuery) Option {
	return func(r *Repository) *Repository {
		r.bigquery = bigquery
		return r
	}
}

func (r *Repository) SUMServiceCostGCP(ctx context.Context, billingTable, billingProject string, from, to time.Time, tz *time.Location, costThreshold float64) ([]domain.GCPServiceCost, error) {
	serviceCost, err := r.bigquery.SUMServiceCostGCP(ctx, billingTable, billingProject, from, to, tz, costThreshold)
	if err != nil {
		return nil, errorz.Errorf("(*bigquery.BigQuery).SUMServiceCostGCP: %w", err)
	}

	return serviceCost, nil
}

func (r *Repository) DailyServiceCostGCP(ctx context.Context, billingTable, billingProject string, from, to time.Time, tz *time.Location, costThreshold float64) ([]domain.GCPServiceCost, error) {
	serviceCost, err := r.bigquery.DailyServiceCostGCP(ctx, billingTable, billingProject, from, to, tz, costThreshold)
	if err != nil {
		return nil, errorz.Errorf("(*bigquery.BigQuery).DailyServiceCostGCP: %w", err)
	}

	return serviceCost, nil
}

func (r *Repository) ServicesOrderBySUMServiceCostGCP(googleCloudPlatformServiceSumCost []domain.GCPServiceCost) []string {
	return slice.Uniq(
		slice.Select(
			slice.Sort(
				googleCloudPlatformServiceSumCost,
				func(a, b domain.GCPServiceCost) bool { return a.Cost > b.Cost },
			),
			func(index int, source domain.GCPServiceCost) (selected string) {
				return source.Service
			},
		),
	)
}

func (r *Repository) DailyServiceCostGCPMapByService(servicesOrderBySUMServiceCostGCP []string, dailyServiceCostGCP []domain.GCPServiceCost) map[string][]domain.GCPServiceCost {
	serviceCost := make(map[string][]domain.GCPServiceCost)
	for _, service := range servicesOrderBySUMServiceCostGCP {
		serviceCost[service] = slice.Filter(dailyServiceCostGCP, func(index int, source domain.GCPServiceCost) bool {
			// nolint: scopelint
			return service == source.Service
		})
	}

	return serviceCost
}
