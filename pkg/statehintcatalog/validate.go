package statehintcatalog

// ValidateSnapshot checks the bounded metadata shape, scope and requested work
// references before a caller narrows annotation bindings. It neither resolves
// ambiguous bindings nor proves permissions, freshness, revision placement or
// event facts. These are the same schema checks used by Adapter.Propose.
func ValidateSnapshot(snapshot Snapshot, scope Scope, refs []string) error {
	if !validRef(scope.WorkspaceRef) || !validRef(scope.OwnerRef) || len(refs) == 0 || len(refs) > MaxWorks {
		return ErrInput
	}
	requested := make(map[string]bool, len(refs))
	for _, ref := range refs {
		if !validRef(ref) || requested[ref] {
			return ErrInput
		}
		requested[ref] = true
	}
	return validateSnapshot(snapshot, scope, requested)
}
