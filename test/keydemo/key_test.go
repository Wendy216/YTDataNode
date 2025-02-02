package keydemo

import (
	ci "github.com/yottachain/YTCrypto"
	"github.com/yottachain/YTDataNode/config"
	"github.com/yottachain/YTDataNode/util"
	"testing"
)

func TestKeySign(t *testing.T) {
	cfg, _ := config.ReadConfig()
	pk, _ := util.Libp2pPkey2eosPkey(cfg.PrivKeyString())
	signstr, _ := ci.Sign(pk, []byte{111, 222})
	if ci.Verify(cfg.PubKey, []byte{111, 222}, signstr) {
		t.Log("pass")
	} else {
		t.Error("err")
	}
}
