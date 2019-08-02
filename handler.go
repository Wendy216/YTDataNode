package node

import (
	"fmt"
	"log"

	"github.com/mr-tron/base58/base58"

	"github.com/yottachain/YTDataNode/message"

	"github.com/golang/protobuf/proto"
	"github.com/yottachain/YTFS/common"
)

// WriteHandler 写入处理器
type WriteHandler struct {
	StorageNode
}

// Handle 获取回调处理函数
func (wh *WriteHandler) Handle(msgData []byte) []byte {
	var msg message.UploadShardRequest
	proto.Unmarshal(msgData, &msg)
	log.Println("超级节点签名:", msg.GetBPDSIGN())
	log.Println("用户签名:", msg.GetUSERSIGN())
	resCode := wh.saveSlice(msg)
	res2client, err := msg.GetResponseToClientByCode(resCode)
	code104, err := msg.GetResponseToClientByCode(104)
	bp := wh.Config().BPList[msg.BPDID]
	if err != nil {
		log.Println("Get res code 2 client fail:", err)
	}
	res2bp, err := msg.GetResponseToBPByCode(resCode, bp.ID, wh.Host().PrivKey())
	if err != nil {
		log.Println("Get res code fail:", err)
	}
	if err != nil {
		log.Println("Get res code 2 bp fail:", err)
	}
	if err = wh.Host().ConnectAddrStrings(bp.ID, bp.Addrs); err != nil {
		log.Println("Connect bp fail", err)
	}
	_, err = wh.Host().SendMsg(bp.ID, "/node/0.0.1", res2bp)
	// 如果报错返回104
	if err != nil {
		return code104
	} else {
		log.Println("return client")
		defer func() {
			err := recover()
			if err != nil {
				log.Println("report to bp error", err)
			}
		}()
		return res2client
	}
}

func (wh *WriteHandler) saveSlice(msg message.UploadShardRequest) int32 {
	// 1. 验证BP签名
	// if ok, err := msg.VerifyBPSIGN(
	// 	// 获取BP公钥
	// 	host.PubKey(wh.Host().Peerstore().PubKey(wh.GetBP(msg.BPDID))),
	// 	wh.Host().ID().Pretty(),
	// ); err != nil || ok == false {
	// 	log.Println(fmt.Errorf("Verify BPSIGN fail:%s", err))
	// 	return 100
	// }
	// 2. 验证数据Hash
	if msg.VerifyVHF(msg.DAT) == false {
		log.Println(fmt.Errorf("Verify VHF fail"))
		return 100
	}
	// 3. 将数据写入YTFS-disk
	var indexKey [32]byte
	copy(indexKey[:], msg.VHF[0:32])
	err := wh.YTFS().Put(common.IndexTableKey(indexKey), msg.DAT)
	if err != nil {
		log.Println(fmt.Errorf("Write data slice fail:%s", err))
		if err.Error() == "YTFS: hash key conflict happens" {
			return 102
		}
		log.Println("数据写入错误error:", err)
		return 101
	}
	log.Println("return msg", 0)
	return 0
}

// DownloadHandler 下载处理器
type DownloadHandler struct {
	StorageNode
}

// Handle 获取处理器
func (dh *DownloadHandler) Handle(msgData []byte) []byte {
	var msg message.DownloadShardRequest
	var indexKey [32]byte
	proto.Unmarshal(msgData, &msg)
	log.Println("get vhf:", base58.Encode(msg.VHF))

	for k, v := range msg.VHF {
		if k >= 32 {
			break
		}
		indexKey[k] = v
	}
	res := message.DownloadShardResponse{}
	resData, err := dh.YTFS().Get(common.IndexTableKey(indexKey))
	if msg.VerifyVHF(resData) {
		log.Println("data verify success")
	}
	if err != nil {
		log.Println("Get data Slice fail:", err)
	}
	res.Data = resData
	resp, err := proto.Marshal(&res)
	if err != nil {
		log.Println("Marshar response data fail:", err)
	}
	log.Println("return msg", 0)
	return append(message.MsgIDDownloadShardResponse.Bytes(), resp...)
}
