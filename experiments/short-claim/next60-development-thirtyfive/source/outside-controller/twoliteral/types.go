// Owned source-only preparation; no original import. All counters describe explicit wrappers.
package twoliteral

import (
	"encoding/json"
	"errors"
)

const (
	Rows                   = 33
	Inputs                 = 11
	MethodsMax             = 61
	CallbacksMax           = 33
	FrameMax               = 384
	EventLineMax           = 4096
	FinalLineMax           = 65536
	AckLineMax             = 256
	JournalMax             = 1634304
	AggregateOutputMax     = 2097152
	BeforeRow              = 0
	BeforeMethod           = 1
	AfterMethod            = 2
	AfterRow               = 3
	Final                  = 4
	BeforeCandidate        = 5
	AfterCandidate         = 6
	BeforeCallback         = 7
	AfterCallback          = 8
	Ready                  = 9
	MethodForEachLine      = 1
	MethodValid            = 2
	MethodParse            = 3
	MethodAppendJSONString = 4
)

var (
	ErrInvalidLine = errors.New("invalid_line")
	ErrInvalidUTF8 = errors.New("invalid_utf8")
	ErrStopped     = errors.New("owned_checkpoint_stopped")
	ErrBounds      = errors.New("owned_bound_unsupported")
)

// Tuple=[reserved,dispatched,returned,panicked]. K.returned is the owned
// callback decision checkpoint, not an assertion that the library consumed it.
type Counts struct {
	Row       [4]int `json:"r"`
	Candidate [4]int `json:"b"`
	Original  [4]int `json:"o"`
	Callback  [4]int `json:"k"`
}
type Primitive struct {
	Type  uint8  `json:"t"`
	Raw   string `json:"r"`
	Index string `json:"i"`
}
type Callback struct {
	PrimitiveKnown bool      `json:"pk"`
	Primitive      Primitive `json:"p"`
	LineKnown      bool      `json:"lk"`
	Line           int       `json:"l"`
	ContinueKnown  bool      `json:"ck"`
	Continue       bool      `json:"c"`
}
type Panic struct {
	Known     bool   `json:"k"`
	Present   bool   `json:"p"`
	TypeKnown bool   `json:"tk"`
	Type      string `json:"t"`
	Scope     string `json:"s"`
}
type ErrorState struct {
	Available        bool   `json:"a"`
	Known            bool   `json:"k"`
	Present          bool   `json:"p"`
	OwnIdentityKnown bool   `json:"ik"`
	OwnIdentity      string `json:"i"`
	TypeKnown        bool   `json:"tk"`
	Type             string `json:"t"`
}
type Got struct {
	NormalKnown               bool        `json:"nk"`
	Normal                    bool        `json:"n"`
	Panic                     Panic       `json:"p"`
	Callbacks                 [3]Callback `json:"c"`
	CallbackN                 int         `json:"cn"`
	MappedPrefix              int         `json:"mp"`
	CallbackListCompleteKnown bool        `json:"clk"`
	CallbackListComplete      bool        `json:"cl"`
	MappingStopped            bool        `json:"ms"`
	StopKnown                 bool        `json:"sk"`
	StopRequested             bool        `json:"sr"`
	Error                     ErrorState  `json:"e"`
	FailingLineKnown          bool        `json:"flk"`
	FailingLine               int         `json:"fl"`
	TerminalKnown             bool        `json:"tk"`
	Terminal                  string      `json:"t"`
	BackingBeforeKnown        bool        `json:"bbk"`
	BackingAfterKnown         bool        `json:"bak"`
	ActiveInputAfterKnown     bool        `json:"ahk"`
	BufferKnown               bool        `json:"bk"`
	ReturnedHex               string      `json:"h"`
	ReturnedNil               bool        `json:"nil"`
	ActiveInputAfterHex       string      `json:"ah"`
	BackingBeforeHex          string      `json:"bb"`
	BackingAfterHex           string      `json:"ba"`
}
type Row struct {
	Ordinal          int             `json:"d"`
	Request          int             `json:"q"`
	Fixture          int             `json:"f"`
	Display          int             `json:"x"`
	Internal         int             `json:"j"`
	RequestID        string          `json:"rid"`
	FixtureID        string          `json:"fid"`
	InputPointer     string          `json:"ip"`
	Input            json.RawMessage `json:"input"`
	Want             json.RawMessage `json:"want"`
	Got              Got             `json:"got"`
	Done             bool            `json:"done"`
	CandidateInvoked bool            `json:"invoked"`
	Truth            *bool           `json:"truth"`
	Role             *string         `json:"role"`
	Weight           *float64        `json:"weight"`
}
type Fact struct {
	Returned               bool       `json:"r"`
	PrimitiveKnown         bool       `json:"pk"`
	Bool                   *bool      `json:"b,omitempty"`
	Primitive              *Primitive `json:"p,omitempty"`
	ReturnedHex            *string    `json:"h,omitempty"`
	ReturnedNil            *bool      `json:"nil,omitempty"`
	BackingHex             *string    `json:"bh,omitempty"`
	CallbackReturnComplete *bool      `json:"cc,omitempty"`
	ErrorChannelAvailable  bool       `json:"ea"`
	Panic                  Panic      `json:"panic"`
	OwnedBoundUnsupported  bool       `json:"ub"`
}
type Result struct {
	Schema                 string    `json:"schema"`
	PlanSHA256             string    `json:"plan_sha256"`
	State                  string    `json:"state"`
	Failure                string    `json:"failure"`
	Completed              int       `json:"completed"`
	Records                [Rows]Row `json:"records"`
	Counts                 Counts    `json:"counts"`
	LastACKSequence        uint32    `json:"last_ack_sequence"`
	LastACKSHA256          string    `json:"last_ack_sha256"`
	OriginalInitialization *int      `json:"original_initialization"`
	OriginalNested         *int      `json:"original_nested"`
	Labels                 int       `json:"labels"`
	Qualified              int       `json:"qualified"`
	NewParents             int       `json:"new_parents"`
	TrainingAuthorized     bool      `json:"training_authorized"`
}
type Frame struct {
	Version           int             `json:"v"`
	Sequence          uint32          `json:"s"`
	Previous          string          `json:"p"`
	Stage             int             `json:"k"`
	Dispatch          int             `json:"d"`
	Request           int             `json:"q"`
	Fixture           int             `json:"f"`
	Display           int             `json:"x"`
	Internal          int             `json:"j"`
	Method            int             `json:"m"`
	Counts            Counts          `json:"n"`
	Row               *Row            `json:"r,omitempty"`
	Fact              *Fact           `json:"a,omitempty"`
	Callback          *Callback       `json:"c,omitempty"`
	Result            json.RawMessage `json:"z,omitempty"`
	DisableEscapeHTML *bool           `json:"html,omitempty"`
}
type Ack struct {
	Version  int    `json:"v"`
	Sequence uint32 `json:"s"`
	SHA256   string `json:"h"`
}
type Checkpoint func(Frame) error

// Ops receives only copied exported primitives. Native type/method getters stay inside binder.
type Ops struct {
	ForEachLine       func(string, func(Primitive) bool)
	Valid             func(string) bool
	Parse             func(string) Primitive
	AppendJSONString  func([]byte, string) []byte
	DisableEscapeHTML bool
}
type Fixture struct {
	Ordinal        int    `json:"ordinal"`
	Request        int    `json:"request_index"`
	Index          int    `json:"fixture_index"`
	RequestID      string `json:"request_id"`
	ID             string `json:"fixture_id"`
	InputReference struct {
		Artifact string `json:"artifact_id"`
		Pointer  string `json:"json_pointer"`
	} `json:"input_reference"`
	Input  json.RawMessage `json:"input"`
	Bytes  int             `json:"input_bytes"`
	SHA256 string          `json:"input_bytes_sha256"`
}
type WantRecord struct {
	Ordinal   int             `json:"fixture_ordinal"`
	RequestID string          `json:"request_id"`
	FixtureID string          `json:"fixture_id"`
	Want      json.RawMessage `json:"Want"`
	Got       json.RawMessage `json:"Got"`
	Truth     json.RawMessage `json:"truth"`
	Role      json.RawMessage `json:"role"`
	Weight    json.RawMessage `json:"weight"`
	Group     json.RawMessage `json:"group"`
}
type LineInput struct {
	Text      string `json:"bytes_utf8"`
	StopAfter *int   `json:"callback_stop_after"`
}
type UTF8Input struct {
	Destination string `json:"destination_hex"`
	Input       string `json:"input_hex"`
}

// Candidate return values stay local until candidate completion. Native baselines
// have no error/failing-line/terminal channels; Availability is explicit.
type LineOutcome struct {
	Err                  error
	ErrorAvailable       bool
	FailingLine          int
	FailingLineAvailable bool
	Terminal             string
	TerminalAvailable    bool
}
type AppendOutcome struct {
	Bytes          []byte
	Err            error
	ErrorAvailable bool
}
