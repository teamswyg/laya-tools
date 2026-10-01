# From 24 to 2,400: reading golden-set counts and evidence

[한국어](golden-set-scale.ko.md)

See the [2,400-request acquisition plan](golden-set-acquisition.en.md) for stages, provenance, task diversity, and execution budgets. Targets and acquired counts remain separate.

Twenty-four requests can check an early mechanism, but provide too little evidence for practical performance. For the project's next performance claims, the target is **at least 2,400 distinct final requests in each evaluation domain where a claim is asserted**. File retrieval, actual model routing, repository selection and decomposition each need their own ground truth and outcomes. File-retrieval requests cannot fill the sample requirement for another domain.

The number 2,400 is a project lower bound, not a universal statistical guarantee. In a simple accuracy rate, changing one case changes the result by about **4.17 percentage points** out of 24, versus **0.0417 percentage points** out of 2,400. This illustrates sample size; it does not fix bias or incorrect labels. Rare under-routing, repository-specific failures and Korean/English differences require enough cases within those slices as well.

## Counts established so far

| Evidence | Established count | Meaning and limitation |
|---|---:|---|
| Synthetic semantic-retrieval pilot | 24 final-probe requests | A one-author transfer probe with limited grammar and subjects. Already observed; development data for subsequent work. |
| Difficulty development fixtures | 36 tasks | Compare classification with author-assigned expected tiers. These are not ground truth for downstream model success. |
| Public candidates for actual coding comparisons | 7 candidates, two code families, one repository; 8 cumulative attempt records on three requests | Preserve plan51 rejection/support-error records. Plan52 accepted six closures, five with zero exit/complete usage and one timeout/unknown usage. One new behavioral contract has no model run. Not representative completion rates or training labels. |
| CoSQA relevance evaluation | 2,400 query/code pairs in 724 connected groups | Actually scored, without establishing useful improvement over BM25. Evaluates pair relevance, not candidate retrieval or routing. |
| Experiment 46 training | 4,456 eligible requests | Development requests used to fit coefficients on 16 numeric features. |
| Experiment 46 validation | 2,599 eligible requests | Used to select epoch, penalty and threshold. Not final evaluation. |
| File-retrieval protected final | 2,402 requests | A protected partition whose outcomes have not been read or scored. This count does not establish final performance or every request's evaluation eligibility. |

See the [24-case synthetic probe](../experiments/semantic-learning/README.en.md), [36 difficulty fixtures](layout-and-goldens.en.md), [2,400-pair evaluation](../experiments/semantic-scale/PAIR-01.en.md), [scale and partition principles](../experiments/semantic-scale/README.en.md), and [experiment 46](../experiments/path-cost-claim/RESULTS-46.en.md). Semantic-scale's 2,400-pair evaluation and experiment 46's 2,402 protected final requests are separate data. Protected final counts assigned membership, not 2,402 ready, eligible goldens. Source and cost eligibility still need separate verification before future evaluation or release.

Despite its larger count, experiment 46 **failed the existing 5% page-reduction gate**. Always using the helper reduced pages on 116 validation requests, increased them on 283, and tied on 2,200. With penalty zero, only 734 training rows and 399 validation rows had nonzero weights. These differ from the 4,456/2,599 input-row counts and do not apply unchanged to other penalties. A large collection does not necessarily contain sufficient signal to distinguish benefit from harm. No new policy was activated or new weights released.

Expected tiers in the [public task candidates](../benchmarks/training/public-task-candidates.json) remain hypotheses until actual model outcomes provide evidence. Record the requested model/reasoning settings, router proposal and settings actually applied separately. Report unobserved usage as unknown rather than zero.

[Experiment 50](../experiments/task-outcomes/RESULTS-50.en.md) was refused before coding-model launch. Separately, [experiment 51](../experiments/task-outcomes/RESULTS-51.en.md) recorded two owned CLI attempts on one comment task. The first had complete usage but failed its exact-change contract; the second returned an unsupported-profile error with unknown usage. Startup-parser incompatibility stopped the remaining four planned entries. Neither record is a training label, and profile-support failure is not model capability failure. Historical [authored record 49](../experiments/task-outcomes/fixtures-49.json), with zero actual model outcomes, remains preserved.

[Separate experiment52](../experiments/task-outcomes/RESULTS-52.en.md) froze executable requested profiles and recorded six attempts on three existing development requests. Both comment pairs now have acceptance/usage comparisons; Sol6 on the behavioral request reached its deadline after closure acceptance, leaving whole cost unknown. Repeating a request previously used in51 does not add a unique request. Three cumulative requests remain far below the2,400 final target.

## Final requests and ground truth needed in each domain

The following table gives **future targets and required evidence**, not measured collection counts or already verified labels.

| Evaluation domain | Final target | Ground truth, evidence and outcomes |
|---|---:|---|
| Search hints and candidate ordering | At least 2,400 distinct retrieval requests | Fixed candidates and acceptable relevant evidence/files; ranks, omissions, cost and quality. Practical utility claims also need independent verification after sufficient evidence is found. |
| Actual model routing | At least 2,400 distinct task requests | Run the compared models under matched conditions and verify success through independent tests/acceptance criteria. The acceptable successful-model set and total cost support the labels. An author's fast/standard/strong tier is an initial hypothesis. |
| Repository selection | At least 2,400 distinct selection requests | Freeze permissions and candidate scope at request time, the acceptable set of repositories that can satisfy the request, and abstention conditions when no answer or access exists. |
| Task decomposition | At least 2,400 distinct parent-task requests | Verify acceptable plans for single, parallel or dependent execution; dependencies and interfaces; correctness after merging; and completion criteria for the entire parent task. |

**Expected task tier**, **proxy search cost** and **verified downstream utility** are different evidence. Experiment 46 counts pages of 20 candidates until the first target file. It does not measure actual LLM tokens, fees or completion time. Model routing must record every attempt and total calls, usage, time and cost, including failures, retries, escalation, abstention and verification. Decomposition must include merging/verification cost and the complete parent outcome, rather than stopping at subtask counts or individual successes.

## Separation still matters at larger scale

Running one request with two seeds or several epochs, models or settings still counts as one request. Group translations, paraphrases and cases derived from shared code or templates. Keep a parent task and its child/sibling tasks in one partition. Tasks from the same repository can also be dependent. Report total requests alongside unique sources, repositories and groups; do not describe every correlated request as an independent sample.

For new collections, freeze grouped training, selection validation, calibration and final partitions **before looking at scores**. Claims about unseen repositories require holding out whole repositories. Freeze data hashes, model/design-selection procedures, cost/quality metrics, failure/out-of-scope/missing treatment and gates in advance. Final data used to change a design becomes development data. Renaming observed validation does not make it final evidence.

A proposed future distribution should cover direct wording and paraphrases, short and long requests, ambiguous or no-answer cases, several repositories, languages and change scopes, and rare costly failures. Specify slice targets and scope first, then report actual support, missing cases and uncertainty. Do not present a proposed distribution as measured counts or turn expected tiers into verified success labels. Existing partitions and gates remain unchanged, with no new fitting or final scoring. The separate development pilots' eight cumulative CLI attempts on three requests do not replace the target of 2,400 independent final requests or establish savings.

The [official SWE-bench evaluation guide](https://www.swebench.com/SWE-bench/guides/evaluation/) is a methodological reference for checking resolution through repository tests after applying a patch. [SWE-bench Goes Live!](https://arxiv.org/abs/2505.23419) is a reference for evaluation design using recent tasks and diverse repositories. Citing them does not adopt their data, approve usage/redistribution rights, or establish that this project's data are free of pretraining contamination.
