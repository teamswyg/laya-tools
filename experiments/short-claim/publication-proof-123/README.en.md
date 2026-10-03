# Actual publication and storage verification

[PR122](https://github.com/teamswyg/laya-tools/pull/122) passed [all four required CI checks](https://github.com/teamswyg/laya-tools/actions/runs/37106112785) and was bot-merged with equal source/merge trees. Offline30 checks and existing native Laya checks are separate; GPU execution was not verified. [Actual CI record](CI-MERGE122.actual.public.v1.json).

Four [Wiki guide/navigation pages](https://github.com/teamswyg/laya-tools/wiki/Next60-Development-EN) were actually pushed and all remote tracked bytes matched. [Wiki record](WIKI30-PUBLICATION.actual.public.v1.json). [Issue19](https://github.com/teamswyg/laya-tools/issues/19#issuecomment-5963281223) was updated with prior text preserved and full remote readback. [Issue record](ISSUE19-SYNC.actual.public.v14.json).

Four inactive copies were compressed separately and actually restored to temporary files with exact SHA/mode0600/durability checks before cleanup.33,561,848 B became2,072,888 B, saving **31,488,960 B before control records**. A nonauthor independently checked full gzip EOF/CRC/SHA, actual absence and unchanged16 older archives. See [review](storage-review/REVIEW.v1.json) and [limits](storage-review/README.en.md). These are disk-storage figures, not model RAM/GPU/speed gains. The saved census is a historical non-atomic sample; fresh admission must keep the existing512MiB cap. Future original-path recovery changes inode/mtime and needs a new metadata freeze.

Actual data30/88 and HF30 remain unchanged. [Three next tasks](../next60-native-three-preparation/README.en.md) have prospectively frozen inputs, Wants, captions and order, with zero new executions/qualification/Fit/models. Model weights, private paths, original bodies and raw journals are excluded.

[한국어](README.ko.md)
