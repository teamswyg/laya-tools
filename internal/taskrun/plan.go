package taskrun

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/teamswyg/laya-tools/internal/taskverify"
	"github.com/teamswyg/laya-tools/pkg/taskoutcome"
	"strings"
)

const MaxPlanBytes = 64 << 10

type planEvaluationRecipe struct {
	TaskID          string `json:"task"`
	PromptSHA256    string `json:"prompt_sha256"`
	RecipeSHA256    string `json:"recipe_sha256"`
	ModuleSHA256    string `json:"module_sha256"`
	LanguageVersion string `json:"language_version"`
}

type plan struct {
	EvaluationRecipes []planEvaluationRecipe `json:"evaluation_recipes,omitempty"`
	GoToolchain       json.RawMessage        `json:"go_toolchain,omitempty"`
	Schema            string                 `json:"schema"`
	Status            string                 `json:"status"`
	PublicBase        string                 `json:"public_task_base_revision"`
	CLI               struct {
		Version string `json:"version"`
		Hash    string `json:"binary_sha256"`
	} `json:"cli"`
	Profiles []struct {
		ID        string `json:"id"`
		Model     string `json:"requested_model"`
		Reasoning string `json:"requested_reasoning"`
	} `json:"profiles"`
	Tasks []struct {
		ID   string `json:"id"`
		Spec string `json:"spec_sha256"`
	} `json:"tasks"`
	Attempts []struct {
		Ordinal int    `json:"ordinal"`
		Task    string `json:"task"`
		Profile string `json:"profile"`
	} `json:"ordered_attempts"`
	Execution struct {
		Max         int  `json:"max_codex_cli_invocations"`
		Concurrency int  `json:"concurrency"`
		Timeout     int  `json:"model_attempt_timeout_seconds"`
		Verify      int  `json:"independent_verification_timeout_seconds"`
		Stdout      int  `json:"stdout_max_bytes"`
		Stderr      int  `json:"stderr_max_bytes"`
		Fresh       bool `json:"fresh_public_workspace_per_attempt"`
		Retries     bool `json:"executor_retries_resume_fallback"`
	} `json:"execution"`
}

func validatePlan(req Request, toolchain trustedToolchain) (string, error) {
	root, e := os.OpenRoot(filepath.Dir(req.PlanFile))
	if e != nil {
		return "", Error("plan_unavailable")
	}
	defer root.Close()
	b, e := readRootFile(root, filepath.Base(req.PlanFile), MaxPlanBytes)
	if e != nil {
		return "", Error("plan_unavailable")
	}
	if digest(b) != req.PlanSHA256 {
		return "", Error("plan_pin_mismatch")
	}
	var p plan
	if json.Unmarshal(b, &p) != nil {
		return "", Error("invalid_plan")
	}
	if len(p.GoToolchain) != 0 {
		var pins struct {
			Version string `json:"version"`
			Hash    string `json:"binary_sha256"`
		}
		if json.Unmarshal(p.GoToolchain, &pins) != nil || pins.Version == "" || pins.Hash == "" || pins.Version != toolchain.version || pins.Hash != toolchain.hash {
			return "", Error("planned_go_toolchain_mismatch")
		}
	}
	if p.Schema != "riido-task-outcome-pilot-plan-v1" || p.Status != "precommitted_development_pilot_not_final_routing_evaluation" || p.CLI.Hash != req.ExpectedCLIHash || p.CLI.Version != req.ExpectedCLIVersion {
		return "", Error("plan_pin_mismatch")
	}
	if len(p.Attempts) == 0 || len(p.Attempts) > 32 || p.Execution.Max != len(p.Attempts) || p.Execution.Concurrency != 1 || p.Execution.Retries || !p.Execution.Fresh || p.Execution.Timeout <= 0 || time.Duration(p.Execution.Timeout)*time.Second != req.Timeout || p.Execution.Verify != 45 || p.Execution.Stdout != 64<<20 || p.Execution.Stderr != MaxStderrBytes || len(p.Tasks) == 0 || len(p.Tasks) > 32 || len(p.Profiles) == 0 || len(p.Profiles) > 16 {
		return "", Error("invalid_plan_limits")
	}
	for i, task := range p.Tasks {
		spec, e := taskverify.TaskSpec(task.ID)
		if e != nil || p.PublicBase != spec.BaseRevision || digestJSON(spec) != task.Spec {
			return "", Error("plan_task_spec_mismatch")
		}
		if e := validatePlannedRecipe(task.ID, spec, p, toolchain); e != nil {
			return "", e
		}
		for _, prior := range p.Tasks[:i] {
			if prior.ID == task.ID {
				return "", Error("duplicate_plan_task")
			}
		}
	}
	for i, pin := range p.EvaluationRecipes {
		found := false
		for _, task := range p.Tasks {
			found = found || task.ID == pin.TaskID
		}
		if !found {
			return "", Error("unknown_planned_recipe")
		}
		for _, prior := range p.EvaluationRecipes[:i] {
			if prior.TaskID == pin.TaskID {
				return "", Error("duplicate_planned_recipe")
			}
		}
	}
	for i, profile := range p.Profiles {
		if profile.ID == "" || profile.Model == "" || profile.Reasoning == "" {
			return "", Error("invalid_plan_profile")
		}
		if _, err := taskoutcome.Summarize(strings.NewReader(""), taskoutcome.Metadata{PublicTaskLabel: profile.ID, RequestedModel: profile.Model, RequestedReasoning: profile.Reasoning}); err != nil {
			return "", Error("invalid_plan_profile")
		}
		for _, prior := range p.Profiles[:i] {
			if prior.ID == profile.ID {
				return "", Error("duplicate_plan_profile")
			}
		}
	}
	if req.AttemptOrdinal < 1 || req.AttemptOrdinal > len(p.Attempts) {
		return "", Error("invalid_attempt_ordinal")
	}
	selectedProfile := ""
	for i, a := range p.Attempts {
		if a.Ordinal != i+1 {
			return "", Error("invalid_plan_order")
		}
		foundTask, foundProfile := false, false
		for _, task := range p.Tasks {
			if task.ID == a.Task {
				foundTask = true
				if a.Ordinal == req.AttemptOrdinal && (task.ID != req.TaskID || task.Spec != req.ExpectedSpecSHA256) {
					return "", Error("planned_attempt_mismatch")
				}
			}
		}
		for _, profile := range p.Profiles {
			if profile.ID == a.Profile {
				foundProfile = true
				if a.Ordinal == req.AttemptOrdinal {
					if profile.Model != req.Model || profile.Reasoning != req.Reasoning {
						return "", Error("planned_attempt_mismatch")
					}
					selectedProfile = profile.ID
				}
			}
		}
		if !foundTask || !foundProfile {
			return "", Error("invalid_plan_reference")
		}
	}
	return selectedProfile, nil
}

func validatePlannedRecipe(id string, spec taskverify.Spec, p plan, toolchain trustedToolchain) error {
	var selected *planEvaluationRecipe
	for i := range p.EvaluationRecipes {
		if p.EvaluationRecipes[i].TaskID == id {
			selected = &p.EvaluationRecipes[i]
		}
	}
	if spec.EvaluationRecipeSHA256 == "" && selected == nil {
		return nil
	}
	if selected == nil || len(p.GoToolchain) == 0 || toolchain.version == "" || toolchain.hash == "" {
		return Error("planned_evaluation_recipe_required")
	}
	recipe, err := taskverify.TaskEvaluationRecipe(id)
	if err != nil || selected.PromptSHA256 != digest([]byte(spec.Prompt)) || selected.RecipeSHA256 != recipe.RecipeSHA256 || selected.ModuleSHA256 != recipe.ModuleSHA256 || selected.LanguageVersion != recipe.LanguageVersion || spec.EvaluationRecipeSHA256 != "" && spec.EvaluationRecipeSHA256 != recipe.RecipeSHA256 {
		return Error("planned_evaluation_recipe_mismatch")
	}
	return nil
}
