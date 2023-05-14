package filters

import (
	"context"
	"log"
	"math/big"
	"reflect"
	"unsafe"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func initEVMClient(RPCAddr string) (*ethclient.Client, error) {
	log.Println("[InitEVMClient] Connecting to: ", RPCAddr)
	return ethclient.Dial(RPCAddr)
}

type GetUniswapV2AddrsFromBlockResp struct {
	MetaData hexutil.Bytes
	Reason   error
}

type UniswapV2LpFilterPacket struct {
	CurrentBlockNo   int64
	CurrentBlockHash common.Hash
	CatchupBlockNo   int64
	CatchupState     bool
	UniswapV2Lps     map[common.Address]GetUniswapV2AddrsFromBlockResp
	TimeTaken        int64
}

func getRPCCl(cl *ethclient.Client) *rpc.Client {
	var clientValue = reflect.ValueOf(cl).Elem()
	fieldStruct := clientValue.FieldByName("c")
	clientPointer := reflect.NewAt(fieldStruct.Type(), unsafe.Pointer(fieldStruct.UnsafeAddr())).Elem()
	rpcClient, _ := clientPointer.Interface().(*rpc.Client)
	return rpcClient
}

func main() {
	ctx := context.Background()
	_cl, err := initEVMClient("/mnt/data/bsc/bsc_ipc")
	if err != nil {
		log.Fatal("[main] {initEVMClient} unable to connect to client, err: ", err.Error())
	}

	// Client check
	_block, err := _cl.BlockNumber(ctx)
	if err != nil {
		log.Fatal("[mai] {initEVMClient} unable to connect to get latest block, err: ", err.Error())
	} else {
		log.Printf("[InitEVMClient] Latest Block No: %+v", _block)
	}

	_rpcCl := getRPCCl(_cl) //get RPC client.

	_sub := make(chan UniswapV2LpFilterPacket)
	_, err = _rpcCl.EthSubscribe(
		context.Background(), _sub, "uniswapV2LpFilter", "0x"+big.NewInt(int64(_block)-1000).Text(16),
	)
	if err != nil {
		log.Fatal("failed to subscribe to uniswapV2LpFilter, err: ", err.Error())
	}

	_avg_ := int64(0)
	_sum_ := int64(0)
	_count := int64(0)
	for {
		select {
		case dat := <-_sub:
			_count += 1
			_sum_ += dat.TimeTaken
			_avg_ = _sum_ / _count
			log.Printf("{streamUnisw} Number of Lps: %d Catchup: %+v Time Taken: %d Average Time Taken: %d", len(dat.UniswapV2Lps), dat.CatchupState, dat.TimeTaken, _avg_)
		}
	}
}