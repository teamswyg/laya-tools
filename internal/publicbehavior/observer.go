// Package publicbehavior observes a closed set of pinned public library APIs.
// It is a development evidence adapter, not a router or a correctness oracle.
package publicbehavior

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"strconv"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/publicbehavior/testdata/upstream/doublestar"
	"github.com/teamswyg/laya-tools/internal/publicbehavior/testdata/upstream/semver"
)

const (
	OperationStrictParse  = "strict_parse"
	OperationCompare      = "compare"
	OperationGlobMatch    = "glob_match"
	OperationGlobValidate = "glob_validate"

	MaxOperationBytes = 32
	MaxInputBytes     = 512

	ErrorNone                 = ""
	ErrorUnsupportedOperation = "unsupported_operation"
	ErrorInputBounds          = "input_bounds"
	ErrorPanic                = "panic"
	ErrorUnclassified         = "unclassified_error"
	ErrorSemverEmpty          = "semver_empty"
	ErrorSemverInvalid        = "semver_invalid"
	ErrorSemverCharacters     = "semver_invalid_characters"
	ErrorSemverSegmentZero    = "semver_segment_starts_zero"
	ErrorSemverMetadata       = "semver_invalid_metadata"
	ErrorSemverPrerelease     = "semver_invalid_prerelease"
	ErrorStrconvNum           = "strconv_num_error"
	ErrorStrconvSyntax        = "strconv_syntax"
	ErrorStrconvRange         = "strconv_range"
	ErrorGlobBadPattern       = "glob_bad_pattern"
)

// Input contains raw strings, never executable code or a function name supplied
// to a dynamic registry. Both strings must be valid UTF-8 and at most 512 bytes;
// even unused Right values are checked before any upstream API is called.
type Input struct {
	Operation string `json:"operation"`
	Left      string `json:"left"`
	Right     string `json:"right"`
}

// VersionObservation preserves one StrictNewVersion call. Called and Returned
// distinguish an unused slot or a panic from a returned nil version. Errors and
// NumError fields are observations of the returned error, not acceptance labels.
type VersionObservation struct {
	Called       bool   `json:"called"`
	Returned     bool   `json:"returned"`
	NilVersion   bool   `json:"nil_version"`
	ErrorKind    string `json:"error_kind"`
	ErrorMessage string `json:"error_message"`
	ErrorFunc    string `json:"error_func"`
	ErrorNum     string `json:"error_num"`
	ErrorCause   string `json:"error_cause"`
	Major        uint64 `json:"major"`
	Minor        uint64 `json:"minor"`
	Patch        uint64 `json:"patch"`
	Original     string `json:"original"`
	String       string `json:"string"`
	Prerelease   string `json:"prerelease"`
	Metadata     string `json:"metadata"`
}

// Observation is an owned, comparable value. Unsupported operations, rejected
// inputs, panics and unclassified errors cannot supply a contract truth label.
// Strict parsing projects Versions[0] to the top level. Comparison projects the
// left version unless only the right parse fails; Versions preserves both sides.
// UpstreamAPICalls counts only requested entry APIs: StrictNewVersion, Compare,
// Match and ValidatePattern. UpstreamGetterCalls separately counts the adapter's
// Original/String/Major/Minor/Patch/Prerelease/Metadata method calls; neither
// counter counts upstream internal helper calls or contract observations.
type Observation struct {
	Supported           bool                  `json:"supported"`
	RejectedInputBounds bool                  `json:"rejected_input_bounds"`
	Panicked            bool                  `json:"panicked"`
	ErrorKind           string                `json:"error_kind"`
	ErrorMessage        string                `json:"error_message"`
	ErrorFunc           string                `json:"error_func"`
	ErrorNum            string                `json:"error_num"`
	ErrorCause          string                `json:"error_cause"`
	NilVersion          bool                  `json:"nil_version"`
	Major               uint64                `json:"major"`
	Minor               uint64                `json:"minor"`
	Patch               uint64                `json:"patch"`
	Original            string                `json:"original"`
	String              string                `json:"string"`
	Prerelease          string                `json:"prerelease"`
	Metadata            string                `json:"metadata"`
	Versions            [2]VersionObservation `json:"versions"`
	ParseErrorSide      string                `json:"parse_error_side"`
	ComparisonCalled    bool                  `json:"comparison_called"`
	Comparison          int                   `json:"comparison"`
	Matched             bool                  `json:"matched"`
	ValidPattern        bool                  `json:"valid_pattern"`
	UpstreamAPICalls    int                   `json:"upstream_api_calls"`
	UpstreamGetterCalls int                   `json:"upstream_getter_calls"`
}

// Observe calls only the four closed entry APIs. Compare always parses both
// strings separately and calls Compare only after two successful nonnil parses.
// Match does not first validate or rewrite a pattern: ValidatePattern is a
// separate operation so upstream early-return behavior remains observable.
// Rejection flags are set before any candidate call. The unsupported-operation
// error takes precedence when the operation and raw bounds are both invalid.
func Observe(in Input) (out Observation) {
	switch in.Operation {
	case OperationStrictParse, OperationCompare, OperationGlobMatch, OperationGlobValidate:
		out.Supported = true
	}
	out.RejectedInputBounds = len(in.Operation) > MaxOperationBytes ||
		len(in.Left) > MaxInputBytes || len(in.Right) > MaxInputBytes ||
		!utf8.ValidString(in.Operation) || !utf8.ValidString(in.Left) || !utf8.ValidString(in.Right)
	if !out.Supported {
		out.ErrorKind = ErrorUnsupportedOperation
		return out
	}
	if out.RejectedInputBounds {
		out.ErrorKind = ErrorInputBounds
		return out
	}

	completed := false
	defer func() {
		// The completion marker also catches panic(nil) without inspecting or
		// exposing the recovered payload. A panic remains explicitly unknown.
		_ = recover()
		if !completed {
			out.Panicked = true
			out.ErrorKind = ErrorPanic
		}
	}()

	switch in.Operation {
	case OperationStrictParse:
		out.observeVersion(in.Left, 0)
		out.projectVersion(0)
		if out.ErrorKind != ErrorNone {
			out.ParseErrorSide = "left"
		}
	case OperationCompare:
		left := out.observeVersion(in.Left, 0)
		right := out.observeVersion(in.Right, 1)
		out.projectVersion(0)
		if out.Versions[0].ErrorKind != ErrorNone {
			out.ParseErrorSide = "left"
		} else if out.Versions[1].ErrorKind != ErrorNone {
			out.projectVersion(1)
			out.ParseErrorSide = "right"
		}
		if left != nil && right != nil && out.Versions[0].ErrorKind == ErrorNone && out.Versions[1].ErrorKind == ErrorNone {
			out.ComparisonCalled = true
			out.UpstreamAPICalls++
			out.Comparison = left.Compare(right)
		}
	case OperationGlobMatch:
		out.UpstreamAPICalls++
		var err error
		out.Matched, err = doublestar.Match(in.Left, in.Right)
		out.recordError(err)
	case OperationGlobValidate:
		out.UpstreamAPICalls++
		out.ValidPattern = doublestar.ValidatePattern(in.Left)
	}
	completed = true
	return out
}

func (out *Observation) observeVersion(raw string, index int) *semver.Version {
	record := &out.Versions[index]
	record.Called = true
	out.UpstreamAPICalls++
	version, err := semver.StrictNewVersion(raw)
	record.Returned = true
	record.NilVersion = version == nil
	e := classifyError(err)
	record.ErrorKind, record.ErrorMessage = e.kind, e.message
	record.ErrorFunc, record.ErrorNum, record.ErrorCause = e.function, e.number, e.cause
	if version != nil {
		out.UpstreamGetterCalls++
		record.Original = version.Original()
		out.UpstreamGetterCalls++
		record.String = version.String()
		out.UpstreamGetterCalls++
		record.Major = version.Major()
		out.UpstreamGetterCalls++
		record.Minor = version.Minor()
		out.UpstreamGetterCalls++
		record.Patch = version.Patch()
		out.UpstreamGetterCalls++
		record.Prerelease = version.Prerelease()
		out.UpstreamGetterCalls++
		record.Metadata = version.Metadata()
	}
	return version
}

func (out *Observation) projectVersion(index int) {
	record := out.Versions[index]
	out.ErrorKind, out.ErrorMessage = record.ErrorKind, record.ErrorMessage
	out.ErrorFunc, out.ErrorNum, out.ErrorCause = record.ErrorFunc, record.ErrorNum, record.ErrorCause
	out.NilVersion = record.NilVersion
	out.Major, out.Minor, out.Patch = record.Major, record.Minor, record.Patch
	out.Original, out.String = record.Original, record.String
	out.Prerelease, out.Metadata = record.Prerelease, record.Metadata
}

type errorRecord struct {
	kind, message, function, number, cause string
}

func (out *Observation) recordError(err error) {
	e := classifyError(err)
	out.ErrorKind, out.ErrorMessage = e.kind, e.message
	out.ErrorFunc, out.ErrorNum, out.ErrorCause = e.function, e.number, e.cause
}

func classifyError(err error) (record errorRecord) {
	if err == nil {
		return record
	}
	record.message = err.Error()
	switch err {
	case semver.ErrEmptyString:
		record.kind = ErrorSemverEmpty
	case semver.ErrInvalidSemVer:
		record.kind = ErrorSemverInvalid
	case semver.ErrInvalidCharacters:
		record.kind = ErrorSemverCharacters
	case semver.ErrSegmentStartsZero:
		record.kind = ErrorSemverSegmentZero
	case semver.ErrInvalidMetadata:
		record.kind = ErrorSemverMetadata
	case semver.ErrInvalidPrerelease:
		record.kind = ErrorSemverPrerelease
	case doublestar.ErrBadPattern:
		record.kind = ErrorGlobBadPattern
	default:
		if numberError, ok := err.(*strconv.NumError); ok {
			record.kind = ErrorStrconvNum
			record.function, record.number = numberError.Func, numberError.Num
			switch numberError.Err {
			case strconv.ErrSyntax:
				record.cause = ErrorStrconvSyntax
			case strconv.ErrRange:
				record.cause = ErrorStrconvRange
			default:
				record.cause = ErrorUnclassified
			}
		} else {
			record.kind = ErrorUnclassified
		}
	}
	return record
}

// SourceArtifact binds the bytes actually embedded in this compilation. Paths
// are repository-relative. The two upstream module declarations are exact byte
// copies named metadata/upstream.go.mod.txt, so they introduce no nested modules.
type SourceArtifact struct {
	Path   string `json:"path"`
	Bytes  int    `json:"bytes"`
	SHA256 string `json:"sha256"`
}

// Explicit filenames keep the compiled source closure closed. Upstream Go
// sources and both full MIT license files are preserved without modification.
//
//go:embed observer.go upstream-manifest.json
//go:embed testdata/upstream/semver/LICENSE.txt testdata/upstream/semver/metadata/upstream.go.mod.txt
//go:embed testdata/upstream/semver/version.go testdata/upstream/semver/collection.go testdata/upstream/semver/constraints.go testdata/upstream/semver/doc.go
//go:embed testdata/upstream/doublestar/LICENSE testdata/upstream/doublestar/metadata/upstream.go.mod.txt
//go:embed testdata/upstream/doublestar/doublestar.go testdata/upstream/doublestar/glob.go testdata/upstream/doublestar/globoptions.go testdata/upstream/doublestar/globwalk.go
//go:embed testdata/upstream/doublestar/match.go testdata/upstream/doublestar/utils.go testdata/upstream/doublestar/validate.go
var compiledSource embed.FS

// SourceArtifacts returns a fresh fixed array; it exposes no mutable registry.
func SourceArtifacts() (out [17]SourceArtifact) {
	paths := [17]string{
		"observer.go", "upstream-manifest.json",
		"testdata/upstream/semver/LICENSE.txt", "testdata/upstream/semver/metadata/upstream.go.mod.txt",
		"testdata/upstream/semver/version.go", "testdata/upstream/semver/collection.go",
		"testdata/upstream/semver/constraints.go", "testdata/upstream/semver/doc.go",
		"testdata/upstream/doublestar/LICENSE", "testdata/upstream/doublestar/metadata/upstream.go.mod.txt",
		"testdata/upstream/doublestar/doublestar.go", "testdata/upstream/doublestar/glob.go",
		"testdata/upstream/doublestar/globoptions.go", "testdata/upstream/doublestar/globwalk.go",
		"testdata/upstream/doublestar/match.go", "testdata/upstream/doublestar/utils.go",
		"testdata/upstream/doublestar/validate.go",
	}
	for i, path := range paths {
		raw := sourceBytes(path)
		digest := sha256.Sum256(raw)
		out[i] = SourceArtifact{
			Path: "internal/publicbehavior/" + path, Bytes: len(raw),
			SHA256: hex.EncodeToString(digest[:]),
		}
	}
	return out
}

// UpstreamManifest returns an owned copy of the embedded original/packaged path
// mapping, immutable upstream revision pins, byte sizes and SHA-256 digests.
func UpstreamManifest() []byte {
	return sourceBytes("upstream-manifest.json")
}

func sourceBytes(path string) []byte {
	raw, err := compiledSource.ReadFile(path)
	if err != nil {
		panic("public_behavior_source_artifact_missing")
	}
	return raw
}
