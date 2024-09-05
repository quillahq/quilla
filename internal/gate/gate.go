package gate

import (
	"strings"

	"github.com/quilla-hq/quilla/types"
	log "github.com/sirupsen/logrus"
)

type GateType int

const (
	GateTypeNone = iota
	GateTypeJob
)

type GateStatus int

const (
	GateStatusPending = iota
	GateStatusFailed
	GateStatusSucceeded
)

type Gate interface {
	ShouldPass(obj interface{}) bool
	Name() string
	Type() GateType
	Status() GateStatus
}

type NilGate struct{}

func (ng *NilGate) Name() string                    { return "nil gate" }
func (ng *NilGate) Type() GateType                  { return GateTypeNone }
func (ng *NilGate) ShouldPass(obj interface{}) bool { return false }
func (ng *NilGate) Status() GateStatus              { return GateStatusPending }

func GetGateFromLabelsOrAnnotations(identifier string, labels map[string]string, annotations map[string]string) Gate {
	gateName, ok := getGateFromMetadata(labels)
	if ok {
		return GetGate(identifier, gateName, &Options{Secret: getSecretTag(labels)})
	}

	gateName, ok = getGateFromMetadata(annotations)
	if ok {
		return GetGate(identifier, gateName, &Options{Secret: getSecretTag(annotations)})
	}

	return &NilGate{}
}

func getGateFromMetadata(meta map[string]string) (string, bool) {
	gate, ok := meta[types.QuillaGateLabel]
	if ok {
		return gate, true
	}

	return "", false
}

func getSecretTag(meta map[string]string) string {
	secret, ok := meta[types.QuillaGateJobSecret]
	if ok {
		return secret
	}

	return ""
}

type Options struct {
	Secret string
}

func GetGate(identifier string, gateName string, options *Options) Gate {
	switch {
	case strings.HasPrefix(gateName, "job:"):
		g, err := NewJobGate(gateName, options.Secret, identifier)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"gate":  gateName,
			}).Error("failed to parse job gate, check your config")
			return &NilGate{}
		}
		return g
	}

	return &NilGate{}
}
