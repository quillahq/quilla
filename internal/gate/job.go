package gate

import (
	"fmt"
	"strings"

	"github.com/quilla-hq/quilla/types"
	v1 "k8s.io/api/batch/v1"
)

type JobGate struct {
	Image  string
	secret string
	status GateStatus
	gate   types.Gate
}

func NewJobGate(gate string, secret string, identifier string) (*JobGate, error) {
	if strings.Contains(gate, ":") {
		parts := strings.Split(gate, ":")
		if len(parts) == 2 {
			return &JobGate{
				Image:  parts[1],
				secret: secret,
				status: GateStatusPending,
				gate: types.Gate{
					Identifier: identifier,
				},
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid job gate: %s", gate)
}

func (jg *JobGate) Name() string       { return "Job Gate" }
func (jg *JobGate) Type() GateType     { return GateTypeJob }
func (jg *JobGate) Status() GateStatus { return jg.status }

func (jg *JobGate) ShouldPass(obj interface{}) bool {
	job := obj.(*v1.Job)
	if job.Status.Succeeded > 0 {
		jg.status = GateStatusSucceeded
	}

	if job.Status.Failed > 0 {
		jg.status = GateStatusFailed
	}

	return jg.status == GateStatusSucceeded
}
