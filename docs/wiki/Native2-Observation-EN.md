# Actual behavior checks for a small claim model

The intended model supplies cheap hints about which similar candidates may satisfy a request. It needs evidence from actual behavior first. Here, two requests from the existing 20-request draft and nine frozen inputs were observed across five original Go candidates in one execution, yielding 23 observations. No new source family was added.

The CIDR network request was satisfied by IPNet on all five inputs. The simple IP and mask candidates each satisfied 0/5. The alternative hook request was satisfied by OrCompose on 4/4 and Compose on 0/4. Similar names conceal different error stopping and `nil` behavior. A real error with an empty message was also observed.

There were no panics or unknowns. Child OS maximum memory was about 9.11MiB; startup/pin-inclusive wall time was about 0.37s. These are original-code observations, not Laya inference, model accuracy or Codex savings. Tool disk size, Go heap and OS memory describe different quantities.

The 23 observations are not counted as 23 distinct requests. New training-qualified requests remain zero pending separate rights, semantic deduplication, connected groups and roles. Model utility will be tested separately: fewer checks than simple sorting while preserving correct answers. Existing utility-failed models remain inactive.

[Detailed results and measurements](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-native2-actual-observations) · [Original-source review](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-native2-actual-semantic-review) · [한국어](https://github.com/teamswyg/laya-tools/wiki/Native2-Observation-KO)
