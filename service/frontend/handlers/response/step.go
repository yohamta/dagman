package response

import (
	"github.com/dagu-dev/dagu/internal/dag"
	"github.com/dagu-dev/dagu/service/frontend/models"
	"github.com/samber/lo"
)

func ToStepObject(step dag.Step) *models.StepObject {
	so := &models.StepObject{
		Args:        step.Args,
		CmdWithArgs: lo.ToPtr(step.CmdWithArgs),
		Command:     lo.ToPtr(step.Command),
		Depends:     step.Depends,
		Description: lo.ToPtr(step.Description),
		Dir:         lo.ToPtr(step.Dir),
		MailOnError: lo.ToPtr(step.MailOnError),
		Name:        lo.ToPtr(step.Name),
		Output:      lo.ToPtr(step.Output),
		Preconditions: lo.Map(step.Preconditions, func(item *dag.Condition, _ int) *models.Condition {
			return ToCondition(item)
		}),
		RepeatPolicy: ToRepeatPolicy(step.RepeatPolicy),
		Script:       lo.ToPtr(step.Script),
		Variables:    step.Variables,
	}
	if step.SubDAG != nil {
		so.Run = step.SubDAG.Name
		so.Params = step.SubDAG.Params
	}
	return so
}

func ToRepeatPolicy(repeatPolicy dag.RepeatPolicy) *models.RepeatPolicy {
	return &models.RepeatPolicy{
		Repeat:   repeatPolicy.Repeat,
		Interval: int64(repeatPolicy.Interval),
	}
}

func ToCondition(cond *dag.Condition) *models.Condition {
	return &models.Condition{
		Condition: cond.Condition,
		Expected:  cond.Expected,
	}
}
