package response

import (
	"github.com/samber/lo"
	"github.com/yohamta/dagu/internal/constants"
	domain "github.com/yohamta/dagu/internal/models"
	"github.com/yohamta/dagu/internal/scheduler"
	"github.com/yohamta/dagu/service/frontend/models"
	"sort"
	"strings"
)

func ToWorkflowLogResponse(logs []*domain.StatusFile) *models.WorkflowLogResponse {
	statusByName := map[string][]scheduler.NodeStatus{}
	for i, l := range logs {
		for _, node := range l.Status.Nodes {
			addStatusGridItem(statusByName, len(logs), i, node)
		}
	}

	grid := lo.MapToSlice(statusByName, func(k string, v []scheduler.NodeStatus) *models.WorkflowLogGridItem {
		return ToWorkflowLogGridItem(k, v)
	})

	sort.Slice(grid, func(i, c int) bool {
		return strings.Compare(lo.FromPtr(grid[i].Name), lo.FromPtr(grid[c].Name)) <= 0
	})

	hookStatusByName := map[string][]scheduler.NodeStatus{}
	for i, l := range logs {
		if l.Status.OnSuccess != nil {
			addStatusGridItem(hookStatusByName, len(logs), i, l.Status.OnSuccess)
		}
		if l.Status.OnFailure != nil {
			addStatusGridItem(hookStatusByName, len(logs), i, l.Status.OnFailure)
		}
		if l.Status.OnCancel != nil {
			addStatusGridItem(hookStatusByName, len(logs), i, l.Status.OnCancel)
		}
		if l.Status.OnExit != nil {
			addStatusGridItem(hookStatusByName, len(logs), i, l.Status.OnExit)
		}
	}
	for _, k := range []string{constants.OnSuccess, constants.OnFailure, constants.OnCancel, constants.OnExit} {
		if v, ok := hookStatusByName[k]; ok {
			grid = append(grid, ToWorkflowLogGridItem(k, v))
		}
	}

	converted := lo.Map(logs, func(item *domain.StatusFile, _ int) *models.WorkflowStatusFile {
		return ToWorkflowStatusFile(item)
	})

	ret := &models.WorkflowLogResponse{
		Logs:     lo.Reverse(converted),
		GridData: grid,
	}

	return ret
}

func addStatusGridItem(data map[string][]scheduler.NodeStatus, logLen, logIdx int, node *domain.Node) {
	if _, ok := data[node.Name]; !ok {
		data[node.Name] = make([]scheduler.NodeStatus, logLen)
	}
	data[node.Name][logIdx] = node.Status
}

func ToWorkflowStatusFile(status *domain.StatusFile) *models.WorkflowStatusFile {
	return &models.WorkflowStatusFile{
		File:   lo.ToPtr(status.File),
		Status: ToWorkflowStatusDetail(status.Status),
	}
}

func ToWorkflowLogGridItem(name string, vals []scheduler.NodeStatus) *models.WorkflowLogGridItem {
	return &models.WorkflowLogGridItem{
		Name: lo.ToPtr(name),
		Vals: lo.Map(vals, func(item scheduler.NodeStatus, _ int) int64 {
			return int64(item)
		}),
	}
}

func ToWorkflowStepLogResponse(logFile, content string, step *domain.Node) *models.WorkflowStepLogResponse {
	return &models.WorkflowStepLogResponse{
		LogFile: lo.ToPtr(logFile),
		Step:    ToNode(step),
		Content: lo.ToPtr(content),
	}
}

func ToWorkflowSchedulerLogResponse(logFile, content string) *models.WorkflowSchedulerLogResponse {
	return &models.WorkflowSchedulerLogResponse{
		LogFile: lo.ToPtr(logFile),
		Content: lo.ToPtr(content),
	}
}