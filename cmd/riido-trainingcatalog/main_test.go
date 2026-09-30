package main

import "testing"

func TestHistoricalLicenseCandidateNames(t *testing.T) {
	for _, name := range []string{"LICENSE", "LICENSE.txt", "LICENSE-APACHE", "LICENCE", "licence.md", "COPYING", "NOTICE.txt"} {
		if !licenseName(name) {
			t.Fatal(name)
		}
	}
	for _, name := range []string{"README.md", "src/main.go", "license_helper.go", "not-a-license"} {
		if licenseName(name) {
			t.Fatal(name)
		}
	}
	for _, name := range []string{"LICENSE", "licenses", "LICENCE", "licences"} {
		if !licenseDir(name) {
			t.Fatal(name)
		}
	}
	for _, name := range []string{"src", "third_party", "license_helper"} {
		if licenseDir(name) {
			t.Fatal(name)
		}
	}
}
func TestCoverageKeepsFailureDenominatorsAndMaxima(t *testing.T) {
	var total coverage
	add(&total, coverage{Snapshots: 1, Roots: 1, Catalogs: 1, Files: 30, MaxFiles: 30, MaxPathBytes: 100, WithLicenseText: 1, LicenseReferences: 2})
	add(&total, coverage{Snapshots: 1, RootFailures: 1, CatalogFailures: 1})
	add(&total, coverage{Snapshots: 1, Roots: 1, Catalogs: 1, Files: 20, MaxFiles: 20, MaxPathBytes: 200, LicenseBlobFailures: 1})
	if total.Snapshots != 3 || total.Catalogs != 2 || total.RootFailures != 1 || total.CatalogFailures != 1 || total.Files != 50 || total.MaxFiles != 30 || total.MaxPathBytes != 200 || total.WithLicenseText != 1 || total.LicenseBlobFailures != 1 {
		t.Fatalf("%+v", total)
	}
}
