// SPDX-License-Identifier: Apache-2.0
package main

// Frozen Root-adopted public30 rows. Each SHA covers the exact row without LF;
// whole-file SHA includes all LF delimiters. This table assigns no new labels.
var expectedRows = [30]rowSpec{
	{"5d3ca46f281c234079aa154a926af45ff8f66d102edd8f794d0a4ecc6ac64e73", "next60-pflag-native-ipnet", 77, 5, 3, 0},
	{"db3335b73e236e8cd556ed25a50fc0c6e3bce665076913af3134db82bf572033", "next60-mapstructure-native-or", 78, 4, 2, 0},
	{"87fe365475b4beb2286851f9aee91c0227880fe2014eaf681c906f790b9344a1", "next60-humanize-precise-ftoa", 76, 6, 3, 1},
	{"f8c32e274c203fcc0515112649b9ca2f66c2050b3e3202e33e5de4a82c69a9c9", "next60-humanize-strict-comma", 76, 6, 3, 1},
	{"79ebc6bb9d7fe60d8d0410890b2c92baf9aa4d5ef2f77836eece13d52f205c7b", "next60-mapstructure-key-collision", 78, 4, 3, 1},
	{"686d4c99080dab927d0e49c69c88408c0ae29b9e3062b30e3ccf3df745de1924", "next60-pflag-map-snapshot", 77, 4, 3, 1},
	{"e20400225cbeaa190ee0352b1312b8c12be4bb7bb2ed0615d82eaa0d046545f2", "next60-humanize-fractional-bytes", 76, 5, 3, 1},
	{"52bfc56db73b60a735906d8ff19f6fb3a95d4e09b89221b2acdb77230f80f678", "next60-go53-pflag-annotation-ownership", 77, 5, 3, 1},
	{"80f6275da1f2ee452da00ad14eff2d5051b8232de3d75a81e39abf0b9a2402cd", "next60-go53-pflag-bounded-count", 77, 6, 3, 1},
	{"770ee6bae57567cf35a90e277f54086c4c2b4961600e7007ff3b33498d161895", "next60-go53-pflag-single-assignment", 77, 5, 3, 1},
	{"f23cd37a6dd52f2bf87b30a4adba3ef2c5fa10ce3a2edf0fe2821c7cc81a4993", "next60-go53-pflag-sensitive-default", 77, 5, 3, 1},
	{"77d2cc69fb1cbd8d95dde8a7ce2bb0c55408181766f821c119c08fc738fbcfe8", "next60-go53-pflag-deferred-function", 77, 5, 3, 1},
	{"7504b8068c64fd4aad7e2cbe92adfe37c61f949941883e7229d3a687cedeb4b6", "next60-go53-pflag-text-error-value", 77, 5, 3, 1},
	{"7e8338cb074beb3f7486445ab9d83bbc2e6f4745f2ccd7ef1018bd340579ead7", "next60-go53-mapstructure-hooks-errors-is", 78, 5, 3, 1},
	{"99fe1afce4a32fda7cd69238213015cfe4743d8b0103b0e5265d309d3a173266", "next60-go53-mapstructure-remain-conflict", 78, 5, 3, 1},
	{"e94b0153d016b65cbaa7ba3550811ce083558f34dc2d04b160ab4b01889d5224", "next60-go53-mapstructure-tag-precedence", 78, 5, 3, 1},
	{"9d8565c7594d041c82c89a5c585aef5aa8d98c95a535b36503f359f5c23a94bc", "next60-mapstructure-nil-config", 78, 4, 3, 1},
	{"aba904615bc14ef8fd6cfd2f7bf99c1fdff517f831b1020ce748bb6447f8caaf", "next60-cobra-error-writer", 77, 4, 2, 0},
	{"b93495d29f4eaba5fba4c0b9bdc87068f0cd40fabaa5837fb297ad6aebeb4b88", "next60-cobra-required-annotation", 77, 5, 3, 1},
	{"baac350f12a94d94e3dbb10bea0287263eeedd36e0ca95b5354cb9b574e7aadf", "next60-cobra-context-preflight", 77, 4, 3, 1},
	{"844b2591bc6f8d1895f38e4a7451e4a1b035c67ba164bf74bce49a277e4107af", "next60-humanize-bigbytes-precision", 76, 6, 3, 1},
	{"004bab38ec308e21913f394669481d8d5bb53ca48a5954095d4ff91a1c5a6aa1", "go53-pflag-bounded-string-slice", 77, 5, 3, 1},
	{"5dc7f11a9a123d3a0a8268abb5eb0514431d56132c8386e5374a5bc777befe06", "go53-cobra-validator-all-errors", 77, 5, 3, 1},
	{"8bcd0fba2dfe4ab659141995cd8b7e8ef6f2c20842c76f963c3290123ba6feea", "next60-afero-sub-validation", 79, 5, 3, 1},
	{"84b0c340b1bd7ac494e49f0242172402da9d20b457971da0940f4f015add4baa", "go53-retryhttp-request-body-snapshot", 80, 4, 3, 1},
	{"b9c19076c2ba155b70ac5883d110edd712ffed75fee349bda53249a70d5f9df2", "next60-ini-quoted-comments", 81, 5, 3, 1},
	{"dab127895de81e3a73a3df6cbde11a6f89f59107c5609097fe6d1c691419dc26", "next60-ini-delete-index", 81, 5, 3, 1},
	{"7545f6a723e2ae36d3582fd2b3351d3fcd6bad9140a8f4c20b4dda2f2b079f63", "next60-ini-byte-budget", 81, 5, 3, 1},
	{"5d5dd2b476b4d3b3bda2c77b02cbae9082a5d3982518a29ee1f295f9afc57d58", "next60-afero-bounded-read", 79, 5, 3, 1},
	{"c00372045f2fe1c8f48ea3b741794ee32e800e80fa5d4175f83bc3d33e2bcc9b", "next60-afero-exclusive-write", 79, 4, 3, 1},
}
