package taskverify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const evaluationRecipeSchema = "riido-task-evaluation-recipe-v1"
const minimalModuleRecipe = "minimal_offline_module_v1"
const upstreamModuleRecipe = "pinned_upstream_module_v1"

// EvaluationRecipe identifies compiler inputs separately from the toolchain.
// RecipeSHA256 hashes all preceding JSON fields, excluding itself. Neither a
// task/spec hash nor a result is included, so a Spec may pin this without a cycle.
type EvaluationRecipe struct {
	Schema           string `json:"schema"`
	Kind             string `json:"kind"`
	ModulePath       string `json:"module_path"`
	ModuleSHA256     string `json:"module_sha256"`
	LanguageVersion  string `json:"language_version"`
	ToolchainVersion string `json:"toolchain_version"`
	RecipeSHA256     string `json:"recipe_sha256,omitempty"`
}

// These are exact public upstream bytes, not candidate configuration. Their
// hashes must match the immutable source manifest before they enter the sandbox.
const humanizeOriginalModule = "module github.com/dustin/go-humanize\n\ngo 1.21\n"
const uuidOriginalModule = "module github.com/google/uuid\n"

// TaskEvaluationRecipe returns an owned, deterministic identity. Legacy task
// recipes describe their unchanged minimal module; only the two v2 tasks use
// their original module language. This API never reads a candidate filesystem.
func TaskEvaluationRecipe(id string) (EvaluationRecipe, error) {
	module, recipe, err := evaluationModule(id)
	if err != nil {
		return EvaluationRecipe{}, err
	}
	recipe.ModuleSHA256 = digest(module)
	b, err := json.Marshal(recipe)
	if err != nil {
		return EvaluationRecipe{}, fmt.Errorf("invalid_evaluation_recipe")
	}
	recipe.RecipeSHA256 = digest(b)
	return recipe, nil
}

func evaluationModule(id string) ([]byte, EvaluationRecipe, error) {
	r := EvaluationRecipe{Schema: evaluationRecipeSchema, Kind: minimalModuleRecipe,
		LanguageVersion: "1.27.1", ToolchainVersion: "go1.27.1"}
	modulePath := "github.com/teamswyg/laya-tools"
	var definition Definition
	if id == humanizeOrdinalTaskV2 {
		definition = humanizeOrdinalDefinition()
	} else if id == uuidCanonicalTaskV2 {
		definition = uuidCanonicalDefinition()
	} else if d, ok := TaskDefinition(id); ok {
		definition = d
	} else if _, err := TaskSpec(id); err != nil {
		return nil, EvaluationRecipe{}, fmt.Errorf("unknown_task")
	}
	if definition.ID != "" {
		var err error
		modulePath, err = definitionModulePath(definition)
		if err != nil {
			return nil, EvaluationRecipe{}, err
		}
	}
	r.ModulePath = modulePath
	module := []byte("module " + modulePath + "\n\ngo 1.27.1\n")
	if id == humanizeOrdinalTaskV2 || id == uuidCanonicalTaskV2 {
		r.Kind = upstreamModuleRecipe
		if id == humanizeOrdinalTaskV2 {
			module, r.LanguageVersion = []byte(humanizeOriginalModule), "1.21"
		} else {
			module, r.LanguageVersion = []byte(uuidOriginalModule), "1.16"
		}
		if digest(module) != definitionPin(definition, "go.mod") || !validUpstreamModule(module, modulePath, r.LanguageVersion) {
			return nil, EvaluationRecipe{}, fmt.Errorf("invalid_upstream_module_recipe")
		}
	}
	return module, r, nil
}

// Only the two original dependency-free module forms are supported. Requiring
// this narrow grammar rejects require/replace/exclude/toolchain/workspace rules;
// it is intentionally not a general go.mod parser or a dependency resolver.
func validUpstreamModule(module []byte, modulePath, language string) bool {
	fields := strings.Fields(string(module))
	if len(fields) != 2 && len(fields) != 4 || fields[0] != "module" || fields[1] != modulePath {
		return false
	}
	if len(fields) == 2 {
		return language == "1.16"
	}
	return fields[2] == "go" && fields[3] == language
}

func trustedEvaluationModule(id string, files []BaseFile) ([]byte, error) {
	module, recipe, err := evaluationModule(id)
	if err != nil {
		return nil, err
	}
	if recipe.Kind != upstreamModuleRecipe {
		return module, nil
	}
	// Verify's source list already contains owned validated base bytes, but this
	// repeated check keeps direct/internal runner callers from supplying config.
	count := 0
	for _, f := range files {
		if f.Path == "go.mod" {
			count++
			if f.SHA256 != digest(module) || !bytes.Equal(f.Data, module) {
				return nil, fmt.Errorf("upstream_module_pin_mismatch")
			}
		}
	}
	if count != 1 {
		return nil, fmt.Errorf("invalid_upstream_module_manifest")
	}
	return module, nil
}
