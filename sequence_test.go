package gosdk_test

import (
	"encoding/json"

	"github.com/NibiruChain/nibiru/gosdk"
	cmtcoretypes "github.com/cometbft/cometbft/rpc/core/types"
)

func (s *NibiruClientSuite) TestSequenceExpectations() {
	s.T().Log("Get sequence and block")
	_, err := s.network.LatestHeight()
	s.NoError(err)

	// Go to next block
	s.NoError(s.network.WaitForNextBlock())
	// TODO: test: "WaitForNextBlock" should probably return the block height

	accAddr := s.val.Address
	getLatestAccNums := func() gosdk.AccountNumbers {
		accNums, err := s.gosdk.GetAccountNumbers(accAddr.String())
		s.NoError(err)
		return accNums
	}
	seq := getLatestAccNums().Sequence

	s.T().Logf("starting sequence %v should not change from waiting a block", seq)
	s.NoError(s.network.WaitForNextBlock())
	newSeq := getLatestAccNums().Sequence
	s.EqualValues(seq, newSeq)

	s.T().Log("broadcast msg n times, same block, expect sequence += n")
	numTxs := uint64(5)
	seqs := []uint64{}
	txResults := make(map[string]*cmtcoretypes.ResultTx)
	for broadcastCount := uint64(0); broadcastCount < numTxs; broadcastCount++ {
		s.NoError(s.network.WaitForNextBlock())
		from, _, _, msgSend := s.msgSendVars()
		txResp, err := s.gosdk.BroadcastMsgsGrpcWithSeq(
			from,
			seq+broadcastCount,
			msgSend,
		)
		s.NoError(err)
		txHashHex := s.AssertTxResponseSuccess(txResp)

		s.T().Logf("Query for tx %v should fail b/c it's the same block.", broadcastCount)
		txResult, err := s.gosdk.TxByHash(txHashHex)
		jsonBz, _ := json.MarshalIndent(txResp, "", "  ")
		s.Assert().Errorf(err, "txResp: %s", jsonBz)

		txResults[txHashHex] = txResult
		seqs = append(seqs, getLatestAccNums().Sequence)
	}

	s.T().Log("expect sequence += n")
	newNewSeq := getLatestAccNums().Sequence
	txResultsJson, _ := json.MarshalIndent(txResults, "", "  ")
	s.EqualValuesf(int(seq+numTxs-1), int(newNewSeq), "seqs: %v\ntxResults: %s", seqs, txResultsJson)

	s.T().Log("After the blocks are completed, tx queries by hash should work.")
	for times := 0; times < 2; times++ {
		s.NoError(s.network.WaitForNextBlock())
	}

	// increment block
	// expect sequence = ?

	s.T().Log("")

	s.T().Log("")

	s.T().Log("")
}
