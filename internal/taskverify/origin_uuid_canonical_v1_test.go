package taskverify

// Original pinned Git blob identities independently bind the vendored public
// source bytes to the upstream tree; these are not model training labels.
var uuidCanonicalPublicBlobs = []struct{ path, sha string }{
	{"dce.go", "fa820b9d3092b433238026b451eca869387d91fb"},
	{"doc.go", "6c2ab76b1de49b5b4c8b7f9c7a68e9a801f85724"},
	{"go.mod", "fc84cd79d4c79a2ad598d6c5285473aa32b34982"},
	{"hash.go", "d3bc00a31b83b71bc3907d2d135fce059588baac"},
	{"json_test.go", "db2d1ccf7deff28a611f4cf24392f5459a5d3f20"},
	{"LICENSE", "5dc68268d900581915a7bfdc1f2be75cd503dd9e"},
	{"marshal.go", "14bd34072b64d2c9cd0491d00454e593af99e1e2"},
	{"node_js.go", "b2a0bc8711b3dcab80f2b96050b318b5748cff70"},
	{"node_net.go", "0cbbcddbd6e81da612534338c786a4e111bc0a04"},
	{"node.go", "d651a2b0619fa44b7f24cdaf731197e6c1023863"},
	{"null_test.go", "fe0fe8dd20acaed30bf2ef6b21385ebbcb3c8306"},
	{"null.go", "d7fcbf28651697b00add519d8b4402a5ab46ffc2"},
	{"seq_test.go", "4f6c549125afcb12785baee59058dbb6297729d1"},
	{"sql_test.go", "929b10f9c9d65a28785c54d9ff70452c1b81472a"},
	{"sql.go", "2e02ec06c0121d1c82a8b07275091549f9a0c950"},
	{"time_test.go", "46354a4a41031f3c91d6d8d2ce6c3d3732a727c4"},
	{"time.go", "efa4d42f2b5830e4ab6a2eab46ba622cacb9af92"},
	{"util.go", "de4bfb1381b51691d734c526837b9d1b30a0b97e"},
	{"uuid_test.go", "906ecbe7f8e2b89a63836dfb1fdbcb5763c625e3"},
	{"uuid.go", "dc75ceefca17f2497d1ac8d99d366251924d2189"},
	{"version1.go", "463109629eee180e87507373a3e957e72c34f179"},
	{"version4.go", "7697802e4d16b255e7ea22a86071cfbe9af6aa39"},
	{"version6_test.go", "690c09dabef8cbb410065dd64119b8298c891ba5"},
	{"version6.go", "17bbafe0fb6315c88d0c3da29106cc0192feea61"},
	{"version7.go", "3fec67102ec3b0e6e6bb4898abb2cfc5d06eeefd"},
}
