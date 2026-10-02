# Next preparation for public Go development sources

The current result is **7 new repositories and 11 proposed distinct behavior goals**. There are **0 new materialized parents, truth labels or execution-complete contracts**. Original Go API execution, fitting, model calls and protected final reads are also 0. Existing 72/76 development data, truth, roles, masks, provenance and audit signals remain unchanged.

This note corresponds to the [machine-readable proposal](NEXT-DEVELOPMENT-SOURCES.proposal.v1.json). It records immutable official GitHub revisions, license-byte SHA values, minimal package files and proposed checks without reproducing whole source files. **Our labels and truth remain proposals.**

## Retained source references

File counts below cover original non-test package Go files plus license/module assets. They are static closure proposals, not successful builds or offline execution evidence. Each is within the current 32-file limit. README and test-import evidence are recorded separately.

| Source / immutable revision | Original license | Non-test Go / closure files | Different proposed goals |
|---|---|---:|---|
| [joho/godotenv](https://github.com/joho/godotenv/tree/97a2850142438b3c357c1d3439e325181afdcc1d) | [MIT, LICENCE](https://github.com/joho/godotenv/blob/97a2850142438b3c357c1d3439e325181afdcc1d/LICENCE) | 2 / 4 | Explicit-value parsing; sorted escaped serialization |
| [google/go-querystring](https://github.com/google/go-querystring/tree/965d79f2113ea0ff039d29828a08a616a0223d48) | [BSD-3-Clause, LICENSE](https://github.com/google/go-querystring/blob/965d79f2113ea0ff039d29828a08a616a0223d48/LICENSE) | 1 / 4 | Tagged struct to repeated URL values |
| [google/shlex](https://github.com/google/shlex/tree/e7afc7fbc51079733e9468cdfd1efcd7d196cd1d) | [Apache-2.0, COPYING](https://github.com/google/shlex/blob/e7afc7fbc51079733e9468cdfd1efcd7d196cd1d/COPYING) | 1 / 3 | Words with quoting, escapes and EOF behavior |
| [go-logfmt/logfmt](https://github.com/go-logfmt/logfmt/tree/804e98fff868b206344991c57a8182172e5ba41e) | [MIT, LICENSE](https://github.com/go-logfmt/logfmt/blob/804e98fff868b206344991c57a8182172e5ba41e/LICENSE) | 4 / 7 | Record decoding; single key/value encoder state |
| [c2h5oh/datasize](https://github.com/c2h5oh/datasize/tree/aa82cc1e65004e2b59a6e44d26f774ca961b24d8) | [MIT, LICENSE](https://github.com/c2h5oh/datasize/blob/aa82cc1e65004e2b59a6e44d26f774ca961b24d8/LICENSE) | 1 / 3 | Unit parsing/error state; exact-unit formatting |
| [mitchellh/go-wordwrap](https://github.com/mitchellh/go-wordwrap/tree/ecf0936a077a4bd73a1cc2ac5c370f2b55618d62) | [MIT, LICENSE.md](https://github.com/mitchellh/go-wordwrap/blob/ecf0936a077a4bd73a1cc2ac5c370f2b55618d62/LICENSE.md) | 1 / 3 | Whitespace and character-width wrapping |
| [vincent-petithory/dataurl](https://github.com/vincent-petithory/dataurl/tree/d1553a71de50473073e188aa79cebf7f993f20fe) | [MIT, LICENSE](https://github.com/vincent-petithory/dataurl/blob/d1553a71de50473073e188aa79cebf7f993f20fe/LICENSE) | 4 / 5 | Media/byte decoding; percent byte conversion |

None repeats a repository in the historical 8-repository acquisition inventory or the humanize/UUID/semver/glob sources. Different repository names alone do not establish independent author flows or semantic cores.

## First prepare 3 executable contracts

A general description is not a truth table. First bind the complete short request, candidate scope and independently authored finite observations. Compare them with original API outputs under a separate later execution plan.

| Priority / direct API | Independent observation contract to prepare | Hard wrong candidates |
|---|---|---|
| [datasize: UnmarshalText](https://github.com/c2h5oh/datasize/blob/aa82cc1e65004e2b59a6e44d26f774ca961b24d8/datasize.go#L116) | uint64 value, receiver state after failure, NumError category/cause; independent big-integer oracle and fixed literals | Decimal units, overflow wrap, old-state preservation on every error, treating bits as bytes |
| [query: Values](https://github.com/google/go-querystring/blob/965d79f2113ea0ff039d29828a08a616a0223d48/query/encode.go#L125) | Absent versus empty keys, repeated-value order and nested names from fixed primitive structs; no arbitrary callbacks | Keep only the last value, omit zeros regardless of tags, flatten nested names |
| [shlex: Split](https://github.com/google/shlex/blob/e7afc7fbc51079733e9468cdfd1efcd7d196cd1d/shlex.go#L403) | Exact token slices, empty quoted words and completed prefixes before later EOF errors | strings.Fields, removal of empty tokens, discarding prior tokens on any error |

“Parse a size” does not specify overflow success, failure or saturation. “Split like a shell” does not include command execution or full POSIX grammar. Expose these distinctions in both the request and finite truth. Existing limits of 512 bytes and 32 normalized words will be checked later; these draft captions are not claimed to have passed them.

**no_answer** means a supported, unambiguous request for which all offered candidates independently violate the contract. **unknown** means ambiguity or unsupported source/observer behavior prevents that judgment. Do not merge the states. Creating wrong wrappers or changing numeric literals does not create independent parent requests.

## Initial scope and rights limits

- godotenv interpolation can fall back to process environment variables. Start with explicit dollar-free values; exclude `autoload`, file access and `Exec`. Preserve partial maps on syntax errors rather than invent an atomic-zero policy.
- Non-test query/logfmt imports are standard library, but original modules/tests require go-cmp. Pin original go.mod and language semantics; verify the offline recipe before execution. datasize has no original Go directive, and dataurl has no original go.mod.
- logfmt slices may change at the next record, so the observer must copy them immediately. Separate encoder validation failures from writer failures. wordwrap may exceed width for long words; do not replace its behavior with a universal maximum-line-width requirement.
- dataurl identifies borrowed Go `net/url` and `text/template/parser` code. Additional BSD lineage/notice requirements remain unresolved, and lexer goroutine error cleanup is unobserved. A root MIT license alone does not clear all extraction or training rights.
- Retain complete licenses and source copyright headers. Do not manufacture an upstream NOTICE. Source reuse conditions, authorship rights for our captions/literals and future dataset/weight licensing require separate consideration.

## Counting and splitting the expansion

Connect repository revisions, aliases, wrappers, shared helpers, nearvariants and translations as whole families. Union copied semantic cores across repositories before splitting. Distinguish standard-library observer infrastructure from copied behavioral helpers and record the reason. Track original developers separately from our AI-assisted English caption/oracle pipeline. Bilingual documentation is not a Korean model-test cohort.

Advance the first 3 goals into verifiable development contracts, then assess the other 8. Acquire further real goals and sources toward **roughly 128 → 256 → 512 independently meaningful DEVELOPMENT problems**. Repeating literals from these 11 goals cannot fill 128 slots. These targets are not new universal minima for every development fit.

**At least 2,400 fresh protected FINAL requests per domain** remains a separate acquisition/protection target. Public development data cannot be recounted as final data. Seven repository groups do not establish sufficient train/validation/final family diversity. Source/group/role/truth metadata belongs in provenance and supervision; scorer features receive only request and candidate text.

This preparation read official sources and checked cached-byte SHA values and path/line bounds. It executed no original API, benchmark, new model/fit, remote publication or CI. No claim is made that these maintainer reads or this AI collaboration had measured zero cost.
