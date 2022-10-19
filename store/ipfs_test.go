package store_test

//type testdata struct {
//	Name        string `json:"name"`
//	Description string `json:"description"`
//}
//
//const (
//	ipfsNodeAddr = "ipfs.io/ipfs"
//)
//
//func TestIpfsAdd(t *testing.T) {
//	testData := &testdata{
//		Name:        "panacea",
//		Description: "medibloc mainnet",
//	}
//
//	newIpfs := store.NewIpfs(ipfsNodeAddr)
//
//	testDataBz, err := json.Marshal(testData)
//	require.NoError(t, err)
//
//	_, err = newIpfs.Add(testDataBz)
//	require.NoError(t, err)
//
//}
//
//func TestIpfsGet(t *testing.T) {
//
//	newIpfs := store.NewIpfs(ipfsNodeAddr)
//
//	file, err := ioutil.ReadFile("./testdata/test_deal.json")
//	require.NoError(t, err)
//
//	cid, err := newIpfs.Add(file)
//	require.NoError(t, err)
//
//	getStrings, err := newIpfs.Get(cid)
//	require.NoError(t, err)
//
//	var deal types.Deal
//	err = json.Unmarshal(file, &deal)
//	require.NoError(t, err)
//
//	require.Equal(t, deal.DataSchema, getStrings)
//}
