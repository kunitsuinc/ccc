package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/kunitsuinc/ccc/pkg/constz"
	"github.com/kunitsuinc/ccc/pkg/domain"
	"github.com/kunitsuinc/ccc/pkg/errorz"
	"github.com/kunitsuinc/ccc/pkg/log"
	"github.com/kunitsuinc/ccc/pkg/repository"
	"github.com/kunitsuinc/util.go/slice"
	"gonum.org/v1/plot/plotter"
)

var ErrMixedCurrenciesDataSourceIsNotSupported = errors.New("mixed currencies data source is not supported")

type UseCase struct {
	r *repository.Repository
}

type Option func(r *UseCase) *UseCase

func New(opts ...Option) *UseCase {
	u := &UseCase{}

	for _, opt := range opts {
		u = opt(u)
	}

	return u
}

func WithRepository(r *repository.Repository) Option {
	return func(u *UseCase) *UseCase {
		u.r = r
		return u
	}
}

func (u *UseCase) PlotDailyServiceCostGCP(ctx context.Context, target io.Writer, projectID, datasetName, tableName string, from, to time.Time, tz *time.Location, imageFormat string) error {
	sumServiceCostGCP, err := u.r.SUMServiceCostGCP(ctx, projectID, datasetName, tableName, from, to, tz, 0.01)
	if err != nil {
		return errorz.Errorf("(*bigquery.BigQuery).ServiceMonthly: %w", err)
	}
	servicesOrderBySUMServiceCost := u.r.ServicesOrderBySUMServiceCostGCP(sumServiceCostGCP)

	dailyServiceCostGCP, err := u.r.DailyServiceCostGCP(ctx, projectID, datasetName, tableName, from, to, tz, 0.01)
	log.Debugf("%v", dailyServiceCostGCP)
	if err != nil {
		return errorz.Errorf("(*bigquery.BigQuery).ServiceDaily: %w", err)
	}
	currencies := slice.Uniq(slice.Select(dailyServiceCostGCP, func(_ int, s domain.GCPServiceCost) (selected string) { return s.Currency }))
	if len(currencies) != 1 {
		return errorz.Errorf("%s.%s.%s: %v: %w", projectID, datasetName, tableName, currencies, ErrMixedCurrenciesDataSourceIsNotSupported)
	}
	currency := currencies[0]
	dailyServiceCostGCPMapByService := u.r.DailyServiceCostGCPMapByService(servicesOrderBySUMServiceCost, dailyServiceCostGCP)

	dailyServiceCostsForPlot := make(map[string]plotter.Values)
	var xAxisPointsCount int // NOTE: X 軸の数値の数を数える
	for k, v := range dailyServiceCostGCPMapByService {
		dailyServiceCostsForPlot[k] = slice.Select(v, func(_ int, source domain.GCPServiceCost) float64 { return source.Cost })
		if len(v) > xAxisPointsCount {
			xAxisPointsCount = len(v)
		}
	}

	if err := domain.Plot1280x720(
		target,
		fmt.Sprintf("Google Cloud Platform Cost (from %s to %s)", from.Format(constz.DateOnly), to.Format(constz.DateOnly)),
		fmt.Sprintf("Date (%s)", tz.String()),
		currency,
		xAxisPointsCount,
		from,
		to,
		tz,
		servicesOrderBySUMServiceCost,
		dailyServiceCostsForPlot,
		imageFormat,
	); err != nil {
		return errorz.Errorf("domain.Plot1280x720: %w", err)
	}

	return nil
}