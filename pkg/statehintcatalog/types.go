// Package statehintcatalog bridges caller-provided content with bounded,
// read-only catalog observations. It exposes no write or notification API.
package statehintcatalog

import (
	"context"
	"errors"
	"time"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

const (
	MaxWorks    = 64
	MaxLabels   = 128
	MaxStatuses = 32
	MaxEmoji    = 32
)

var (
	ErrInput     = errors.New("statehintcatalog: invalid or oversized request")
	ErrSnapshot  = errors.New("statehintcatalog: invalid or oversized snapshot")
	ErrPredictor = errors.New("statehintcatalog: unavailable or invalid predictor")
)

// Scope references are opaque. Authentication is supplied by the application
// through context and its private reader, never by a field in user JSON.
type Scope struct {
	WorkspaceRef string `json:"workspace_ref"`
	OwnerRef     string `json:"owner_ref"`
}

// Request carries content already available to the caller. Readers need not
// fetch extra documents, messages, titles, or other live content to classify it.
type Request struct {
	WorkRef    string `json:"work_ref"`
	CommandRef string `json:"command_ref"`
	Text       string `json:"text"`
}

type Work struct {
	Ref                    string          `json:"ref"`
	CanonicalVersion       string          `json:"canonical_version,omitempty"`
	State                  statehint.State `json:"state,omitempty"`
	ExistingOpaqueLabels   []string        `json:"existing_opaque_labels"`
	CurrentEmoji           string          `json:"current_emoji,omitempty"`
	StatePlanningAvailable bool            `json:"state_planning_available"`
}

type Label struct {
	Ref            string `json:"ref"`
	Active         bool   `json:"active"`
	GroupContainer bool   `json:"group_container"`
}

// Bindings are explicit application configuration. The bridge does not infer
// a catalog's meaning from label names, ordering, or private native IDs.
type Binding struct {
	Intent    statehint.Intent `json:"intent"`
	LabelRefs []string         `json:"label_refs"`
}

type Status struct {
	Ref       string          `json:"ref"`
	StateType statehint.State `json:"state_type"`
	Active    bool            `json:"active"`
}

type Catalog struct {
	Labels              []Label                    `json:"labels"`
	Bindings            []Binding                  `json:"bindings"`
	Emojis              []statehint.EmojiCandidate `json:"emojis"`
	Statuses            []Status                   `json:"statuses"`
	PreferredStatusRefs map[statehint.State]string `json:"preferred_status_refs"`
}

// Times report observation windows only. Zero times mean unspecified. They do
// not establish atomicity, freshness, current permission, or write authority.
type ReadWindow struct {
	Started   time.Time `json:"started"`
	Completed time.Time `json:"completed"`
}
type ReadTimes struct {
	Works    ReadWindow `json:"works"`
	Labels   ReadWindow `json:"labels"`
	Statuses ReadWindow `json:"statuses"`
}

type Snapshot struct {
	Scope     Scope     `json:"scope"`
	Works     []Work    `json:"works"`
	Catalog   Catalog   `json:"catalog"`
	ReadTimes ReadTimes `json:"read_times"`
}

// SnapshotReader is implemented by the application's trusted read boundary.
// All native IDs, credentials and endpoint details stay behind this interface.
// Separate reads are never promoted to an atomic snapshot by this package.
type SnapshotReader interface {
	Read(context.Context, Scope, []string) (Snapshot, error)
}

type Predictor interface {
	Predict(string) (statehint.Prediction, error)
}
type PredictorFunc func(string) (statehint.Prediction, error)

func (f PredictorFunc) Predict(text string) (statehint.Prediction, error) {
	if f == nil {
		return statehint.Prediction{}, ErrPredictor
	}
	return f(text)
}

type Adapter struct {
	MinConfidence float64
	MinMargin     float64
}

type Proposal struct {
	WorkRef          string                `json:"work_ref"`
	CommandRef       string                `json:"command_ref"`
	CanonicalVersion string                `json:"canonical_version,omitempty"`
	Prediction       *statehint.Prediction `json:"prediction,omitempty"`
	LabelRefs        []string              `json:"label_refs"`
	EmojiCode        string                `json:"emoji_code,omitempty"`
	CurrentState     statehint.State       `json:"current_state,omitempty"`
	NextState        statehint.State       `json:"next_state,omitempty"`
	// This means a shadow Plan can be constructed, never current write authority.
	StatePlanningAvailable bool   `json:"state_planning_available"`
	StateReason            string `json:"state_reason"`
	// Candidate metadata requires separate verified evidence before any future
	// application could consider a state transition; this package never does so.
	CandidateStatusRef string                `json:"candidate_status_ref,omitempty"`
	StateChange        bool                  `json:"state_change"`
	NoOp               bool                  `json:"no_op"`
	MutationExecuted   bool                  `json:"mutation_executed"`
	Plan               *statehint.PlanResult `json:"plan,omitempty"`
}

type MappingIssue struct {
	Reason string           `json:"reason"`
	Ref    string           `json:"ref,omitempty"`
	Intent statehint.Intent `json:"intent,omitempty"`
}

type Result struct {
	Mode               string         `json:"mode"`
	MutationExecuted   bool           `json:"mutation_executed"`
	CatalogConsistency string         `json:"catalog_consistency"`
	Scope              Scope          `json:"scope"`
	ReadTimes          ReadTimes      `json:"read_times"`
	Proposals          []Proposal     `json:"proposals"`
	MappingIssues      []MappingIssue `json:"mapping_issues"`
}
