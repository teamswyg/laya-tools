# Small behavioral claims from public libraries57

The goal is to expand **verifiable hint evidence** about small requirements that code candidates may satisfy. Start with version strings and path patterns. For requests such as ignoring build metadata in version precedence or matching nested directories, a tiny model could suggest which candidate to check. Existing tools retain responsibility for acceptance and truth.

This stage compares original Go functions against independently written expectations for finite inputs. It does not demonstrate general code understanding, router savings or successful training. Do not revise expectations after observing results to manufacture a pass.

## Sources and rights

| Source | Pinned version and revision | Behaviors |
|---|---|---|
| [Masterminds/semver](https://github.com/Masterminds/semver/tree/61fc460d28283a91c53be65c2e0f20b494ac8ad9) | v3.4.0 / `61fc460d28283a91c53be65c2e0f20b494ac8ad9` | Strict parsing, metadata, prerelease precedence |
| [bmatcuk/doublestar](https://github.com/bmatcuk/doublestar/tree/8b690afa33319b0a1869367f594e53977e38bc99) | v4.9.1 / `8b690afa33319b0a1869367f594e53977e38bc99` | Slash/glob, escaping/character matching, pattern validation |

The15 original files total99,035 bytes. Preserve both complete MIT copyright, permission and disclaimer notices. Keep unchanged Go source in audit fixtures. Preserve original module files under metadata without changing the parent module's dependencies. Record byte and SHA mappings between original and packaged paths. Link [SemVer2.0.0](https://semver.org/spec/v2.0.0.html) and author expectations separately.

## Six properties, two source families

1. Strict parsing: preserve original string, numeric components, prerelease and metadata on success; nil and exact error kind on failure. Library error identities are descriptive API contracts, not error names mandated by SemVer.
2. Metadata: distinguish string preservation from having no influence on precedence.
3. Prerelease: cover releases, numeric/text identifiers, dotted prefixes and ASCII case. **Include numeric identifiers larger than uint64**, preserving possible normative discrepancies.
4. Glob: cover whole-name matching, slash boundaries, `*` and component-only `**`. Exclude filesystem operations and OS-dependent `PathMatch`.
5. Escape/character: cover escaped wildcards and `?`, with ASCII and Unicode boundaries stated per input.
6. Validation: observe `ValidatePattern` bool separately from `Match` bool/error. Include first-alternative success in a malformed brace pattern. Distinguish a stronger reject-the-whole-pattern policy from original API behavior as a **policy discrepancy**.

Source reading predicts an issue with large prerelease numeric comparison and early success on malformed braces. These are pre-observation hypotheses. Keep normative, descriptive API and separate policy expectations distinct; do not copy implementation hypotheses into accepted truth. Aliases, helpers, reversed pairs and input variants do not become independent tasks or families.

## Execution boundary

Use Go1.27.1, pure Go audit execution, offline observations and input/file bounds. Original APIs receive raw UTF-8. Never lowercase or strip punctuation from API inputs. Do not automatically fill hint inputs with outcomes or source IDs.

Freeze source, observer, plan and expectations; verify actual binary SHA and input freeze before one official collection. Count property observations separately from requested entry API calls: comparison requires two strict parses and one comparison. Record getter/String/Original calls separately; do not count upstream internal helpers. Preserve errors, panic and unsupported states; incomplete observations do not produce false labels. Post-collection regression replay does not add official samples.

The production `riidolaya` hint path gains no library loading or locks. Fixtures and auditor are a separate maintainer tool. Heap soft limits and execution settings are not measured resource results.

## Subsequent decision

Preserve previous pending/unknown evidence and measurements. New public behavior evidence does not fulfill240 sources, at least15 relationship groups,5% necessary model-free utility or2400 distinct protected-final requests per domain. New fits, paid model calls, final evaluation and weight releases remain0. Reassess training eligibility after diversity and actual checking-work reduction are demonstrated.
