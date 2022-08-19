package panacea_test

// Test for sgxLevelDB function.
// Comment out this test because it can only work in sgx environment.

//func TestSgxLevelDB(t *testing.T) {
//	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
//	require.NoError(t, err)
//	ctx := context.Background()
//
//	queryClient, err := panacea.NewQueryClient(ctx, "panacea-3", "https://rpc.gopanacea.org:443", 99, hash)
//	require.NoError(t, err)
//
//	lightClient := queryClient.LightClient
//	_, err = lightClient.VerifyLightBlockAtHeight(ctx, 1000, time.Now())
//	require.NoError(t, err)
//
//	// get Block info using sgxLevelDB function
//	storedLightBlock, err := lightClient.TrustedLightBlock(1000)
//	require.NoError(t, err)
//	fmt.Println("storedLightBlock at ", storedLightBlock.Height, " : ", storedLightBlock.Hash())
//
//}
