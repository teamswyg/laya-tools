package taskverify

import (
	"slices"
	"testing"
)

func TestOwnedModuleIdentitiesAndOptionalNotice(t *testing.T) {
	for _, id := range []string{keywordGuardTask, eventKeyBoundsTask, humanizeOrdinalTask, uuidCanonicalTask} {
		d, ok := TaskDefinition(id)
		if !ok {
			t.Fatal("missing owned definition")
		}
		module, err := definitionModulePath(d)
		if err != nil || module == "" {
			t.Fatalf("owned module unavailable: %s", id)
		}
		pins, err := AttributionPins(id)
		if err != nil || len(pins) == 0 || pins[0].Path != "LICENSE" {
			t.Fatalf("missing original license: %s", id)
		}
		external := id == humanizeOrdinalTask || id == uuidCanonicalTask
		if external && (len(pins) != 1 || module != d.Packages[0]) || !external && len(pins) != 2 {
			t.Fatalf("fabricated attribution or changed module: %s", id)
		}
		pins[0].SHA256 = "changed caller copy"
		fresh, _ := AttributionPins(id)
		if fresh[0].SHA256 != definitionPin(d, "LICENSE") {
			t.Fatal("caller changed owned attribution")
		}
	}
}

func TestModuleIdentityRefusesRecipeInjectionAndPackageEscape(t *testing.T) {
	base := Definition{ModulePath: "github.com/example/public", Packages: []string{"github.com/example/public/pkg"}}
	if _, err := definitionModulePath(base); err != nil {
		t.Fatal(err)
	}
	for _, module := range []string{"local", "github.com//public", "github.com/example/../public", "github.com/example/public\nreplace x => y", "github.com/example/공개"} {
		d := base
		d.ModulePath = module
		if _, err := definitionModulePath(d); err == nil {
			t.Fatal("unsafe module recipe accepted")
		}
	}
	for _, pkg := range []string{"github.com/example/public-other", "github.com/example/public/../other", "github.com/example/public//pkg", "github.com/example/public/pkg\nreplace", "github.com/example/public/./pkg"} {
		d := base
		d.Packages = []string{pkg}
		if _, err := definitionModulePath(d); err == nil {
			t.Fatal("package escaped owned module recipe")
		}
	}
	base.Packages = nil
	if _, err := definitionModulePath(base); err == nil {
		t.Fatal("empty package manifest accepted")
	}
	d, _ := TaskDefinition(keywordGuardTask)
	d.Packages = slices.Clone(d.Packages)
	if got, err := definitionModulePath(d); err != nil || got != "github.com/teamswyg/laya-tools" {
		t.Fatal("legacy recipe identity changed")
	}
}
