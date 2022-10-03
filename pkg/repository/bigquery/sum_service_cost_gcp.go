// nolint: dupl
package bigquery

import (
	"context"
	"text/template"
	"time"

	"github.com/kunitsuinc/ccc/pkg/constz"
	"github.com/kunitsuinc/ccc/pkg/domain"
	"github.com/kunitsuinc/ccc/pkg/errorz"
	"github.com/kunitsuinc/ccc/pkg/log"
)

type sumServiceCostGCPParameter struct {
	TimeZone          *time.Location
	GCPBillingTable   string
	GCPBillingProject string
	From              string
	To                string
	CostThreshold     float64
}

// nolint: gochecknoglobals
var sumServiceCostGCPTemplate = template.Must(template.New("SUMServiceCostGCP").Parse(`-- SUMServiceCostGCP
SELECT
    service.description AS service,
    ROUND(SUM(cost * 100)) / 100 AS cost,
    currency
FROM
    ` + "`{{ .GCPBillingTable }}`" + `
WHERE
    project.id = '{{ .GCPBillingProject }}'
AND
    DATE(usage_start_time, '{{ .TimeZone }}') >= DATE("{{ .From }}", '{{ .TimeZone }}')
AND
    DATE(usage_start_time, '{{ .TimeZone }}') <= DATE("{{ .To }}", '{{ .TimeZone }}')
AND
    cost >= {{ .CostThreshold }}
GROUP BY
    service, currency
ORDER BY
    cost
DESC
;`))

func (c *BigQuery) SUMServiceCostGCP(ctx context.Context, billingTable, billingProject string, from, to time.Time, tz *time.Location, costThreshold float64) ([]domain.GCPServiceCost, error) {
	q, err := buildQuery(sumServiceCostGCPTemplate, sumServiceCostGCPParameter{
		TimeZone:          tz,
		GCPBillingTable:   billingTable,
		GCPBillingProject: billingProject,
		From:              from.Format(constz.DateOnly),
		To:                to.Format(constz.DateOnly),
		CostThreshold:     costThreshold,
	})
	if err != nil {
		return nil, errorz.Errorf("buildQuery: %w", err)
	}

	log.Debugf("%s", q)

	results, err := query[domain.GCPServiceCost](ctx, c.client, q)
	if err != nil {
		return nil, errorz.Errorf("query: %w", err)
	}

	return results, nil
}
