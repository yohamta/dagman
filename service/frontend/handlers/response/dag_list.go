package response

import (
	"github.com/dagu-dev/dagu/internal/dag"
	"github.com/dagu-dev/dagu/internal/persistence"
	"github.com/dagu-dev/dagu/service/frontend/models"
	"github.com/samber/lo"
)

func ToListDagResponse(
	dagStatusList []*persistence.DAGStatus,
	errs []string,
	hasError bool,
) *models.ListDagsResponse {
	return &models.ListDagsResponse{
		DAGs: lo.Map(dagStatusList, func(item *persistence.DAGStatus, _ int) *models.DagListItem {
			return ToDagListItem(item)
		}),
		Errors:   errs,
		HasError: lo.ToPtr(hasError),
	}
}

func ToDagListItem(s *persistence.DAGStatus) *models.DagListItem {
	return &models.DagListItem{
		Dir:       lo.ToPtr(s.Dir),
		Error:     lo.ToPtr(toErrorText(s.Error)),
		ErrorT:    s.ErrorT,
		File:      lo.ToPtr(s.File),
		Status:    ToDagStatus(s.Status),
		Suspended: lo.ToPtr(s.Suspended),
		DAG:       ToDAG(s.DAG),
	}
}

func ToDAG(dg *dag.DAG) *models.Dag {
	return &models.Dag{
		Name:          lo.ToPtr(dg.Name),
		Group:         lo.ToPtr(dg.Group),
		Description:   lo.ToPtr(dg.Description),
		Params:        dg.Params,
		DefaultParams: lo.ToPtr(dg.DefaultParams),
		Tags:          dg.Tags,
		Schedule: lo.Map(dg.Schedule, func(item dag.Schedule, _ int) *models.Schedule {
			return ToSchedule(item)
		}),
	}
}

func ToSchedule(s dag.Schedule) *models.Schedule {
	return &models.Schedule{
		Expression: lo.ToPtr(s.Expression),
	}
}
