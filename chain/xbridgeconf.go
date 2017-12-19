package chain

import (
	"fmt"
	"strings"
)

const (
	WalletBTC   = "BTC"
	WalletLTC   = "LTC"
	WalletSYS   = "SYS"
	WalletDASH  = "DASH"
	WalletDGB   = "DGB"
	WalletDYN   = "DYN"
	WalletDOGE  = "DOGE"
	WalletPIVX  = "PIVX"
	WalletVIA   = "VIA"
	WalletVTC   = "VTC"
	WalletMUE   = "MUE"
	WalletNMC   = "NMC"
	WalletQTUM  = "QTUM"
	WalletLBC   = "LBC"
	WalletMONA  = "MONA"
	WalletBLOCK = "BLOCK"
	WalletFAIR  = "FAIR"
)

var Wallets = map[string]string{
	WalletBTC:   WalletBTC,
	WalletLTC:   WalletLTC,
	WalletSYS:   WalletSYS,
	WalletDASH:  WalletDASH,
	WalletDGB:   WalletDGB,
	WalletDYN:   WalletDYN,
	WalletDOGE:  WalletDOGE,
	WalletPIVX:  WalletPIVX,
	WalletMONA:  WalletMONA,
	WalletVIA:   WalletVIA,
	WalletVTC:   WalletVTC,
	WalletMUE:   WalletMUE,
	WalletNMC:   WalletNMC,
	WalletQTUM:  WalletQTUM,
	WalletLBC:   WalletLBC,
	WalletBLOCK: WalletBLOCK,
	WalletFAIR:  WalletFAIR,
}

type XWallet struct {
	Name      string
	Address   string
	IP        string
	Port      string
	RPCPort   string
	RPCUser   string
	RPCPass   string
	Container string
	Version   string
	CLI       string
}

func (w XWallet) IsNull() bool {
	return w.Name == ""
}

// SupportsWallet returns true if the wallet is supported.
func SupportsWallet(wallet string) bool {
	_, ok := Wallets[wallet]
	return ok
}

// CreateXWallet returns the default wallet data.
func CreateXWallet(coin, version, address, ip, rpcuser, rpcpass string) XWallet {
	// TODO Update ports
	getCoinVersion := func(repo, ver, defaultVersion string) string {
		if ver == "" {
			ver = defaultVersion
			if ver == "" {
				ver = "latest"
			}
		}
		return fmt.Sprintf("%s:%s", repo, ver)
	}
	switch coin {
	case WalletBTC:
		return XWallet{coin, address, ip, "8333", "8332", rpcuser, rpcpass, getCoinVersion("", version, ""), version, "Unspecified"}
	case WalletLTC:
		return XWallet{coin, address, ip, "9333", "9332", rpcuser, rpcpass, getCoinVersion("blocknetdx/litecoin", version, ""), version, "litecoin-cli"}
	case WalletSYS:
		return XWallet{coin, address, ip, "8369", "8370", rpcuser, rpcpass, getCoinVersion("blocknetdx/syscoin2", version, "2.1.6-snap500644"), version, "syscoin-cli"}
	case WalletDASH:
		return XWallet{coin, address, ip, "", "9998", rpcuser, rpcpass, getCoinVersion("", version, ""), version, "Unspecified"}
	case WalletDGB:
		return XWallet{coin, address, ip, "", "14022", rpcuser, rpcpass, getCoinVersion("", version, ""), version, "Unspecified"}
	case WalletDYN:
		return XWallet{coin, address, ip, "", "31350", rpcuser, rpcpass, getCoinVersion("", version, ""), version, "Unspecified"}
	case WalletDOGE:
		return XWallet{coin, address, ip, "", "22555", rpcuser, rpcpass, getCoinVersion("", version, ""), version, "Unspecified"}
	case WalletPIVX:
		return XWallet{coin, address, ip, "", "51473", rpcuser, rpcpass, getCoinVersion("", version, ""), version, "Unspecified"}
	case WalletVIA:
		return XWallet{coin, address, ip, "", "5222", rpcuser, rpcpass, getCoinVersion("", version, ""), version, "Unspecified"}
	case WalletVTC:
		return XWallet{coin, address, ip, "", "5888", rpcuser, rpcpass, getCoinVersion("", version, ""), version, "Unspecified"}
	case WalletMUE:
		return XWallet{coin, address, ip, "", "29683", rpcuser, rpcpass, getCoinVersion("", version, ""), version, "Unspecified"}
	case WalletNMC:
		return XWallet{coin, address, ip, "", "8336", rpcuser, rpcpass, getCoinVersion("", version, ""), version, "Unspecified"}
	case WalletQTUM:
		return XWallet{coin, address, ip, "", "3889", rpcuser, rpcpass, getCoinVersion("", version, ""), version, "Unspecified"}
	case WalletLBC:
		return XWallet{coin, address, ip, "", "9245", rpcuser, rpcpass, getCoinVersion("", version, ""), version, "Unspecified"}
	case WalletMONA:
		return XWallet{coin, address, ip, "9401", "9402", rpcuser, rpcpass, getCoinVersion("blocknetdx/monacoin", version, "0.14.2-snap1193272"), version, "monacoin-cli"}
	case WalletBLOCK:
		return XWallet{coin, address, ip, "41412", "41414", rpcuser, rpcpass, getCoinVersion("blocknetdx/servicenode", version, ""), version, "blocknetdx-cli"}
	case WalletFAIR:
		return XWallet{coin, address, ip, "", "40405", rpcuser, rpcpass, getCoinVersion("", version, ""), version, "Unspecified"}
	}
	return XWallet{}
}

// DefaultXConfig returns the default xbridge config for the specified coin.
func DefaultXConfig(coin, version, address, ip, rpcuser, rpcpass string) string {
	wallet := CreateXWallet(coin, version, address, ip, rpcuser, rpcpass)
	switch coin {
	case WalletBTC:
		return BTC(wallet)
	case WalletLTC:
		return LTC(wallet)
	case WalletSYS:
		return SYS(wallet)
	case WalletDASH:
		return DASH(wallet)
	case WalletDGB:
		return DGB(wallet)
	case WalletDYN:
		return DYN(wallet)
	case WalletDOGE:
		return DOGE(wallet)
	case WalletPIVX:
		return PIVX(wallet)
	case WalletVIA:
		return VIA(wallet)
	case WalletVTC:
		return VTC(wallet)
	case WalletMUE:
		return MUE(wallet)
	case WalletNMC:
		return NMC(wallet)
	case WalletQTUM:
		return QTUM(wallet)
	case WalletLBC:
		return LBC(wallet)
	case WalletMONA:
		return MONA(wallet)
	case WalletBLOCK:
		return BLOCK(wallet)
	case WalletFAIR:
		return FAIR(wallet)
	}

	return ""
}

// MAIN returns the main config section.
func MAIN(wallets []string) string {
	return `[Main]
ExchangeWallets=` + strings.Join(wallets, ",") + `
FullLog=true
LogPath=/var/log/xbridge.log
ExchangeTax=300

[RPC]
Enable=false
UserName=
Password=
UseSSL=false
Port=9898

`
}

// BTC bitcoin
func BTC(wallet XWallet) string {
	return `[BTC]
Title=Bitcoin
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.Port + ` # 8332
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=0
ScriptPrefix=5
SecretPrefix=128
COIN=100000000
MinimumAmount=0
TxVersion=2
DustAmount=0
CreateTxMethod=BTC
MinTxFee=27000
BlockTime=600
GetNewKeySupported=false
ImportWithNoScanSupported=false
FeePerByte=105
Confirmations=0
`
}

// LTC litecoin
func LTC(wallet XWallet) string {
	return `[LTC]
Title=Litecoin
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 9332
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=48
Scrwallet.IPtPrefix=5
SecretPrefix=176
COIN=100000000
MinimumAmount=0
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=false
ImportWithNoScanSupported=true
FeePerByte=110
MinTxFee=60000
TxVersion=1
BlockTime=60
Confirmations=0
`
}

// SYS syscoin2
func SYS(wallet XWallet) string {
	return `[SYS]
Title=SysCoin2
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 8370
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=0
Scrwallet.IPtPrefix=5
SecretPrefix=128
COIN=100000000
MinimumAmount=0
TxVersion=1
DustAmount=0
CreateTxMethod=BTC
MinTxFee=60000
BlockTime=60
GetNewKeySupported=false
ImportWithNoScanSupported=false
FeePerByte=100
Confirmations=0
`
}

// DASH dash
func DASH(wallet XWallet) string {
	return `[DASH]
Title=Dash
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 9998
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=76
Scrwallet.IPtPrefix=16
SecretPrefix=204
COIN=100000000
MinimumAmount=0
TxVersion=1
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=false
ImportWithNoScanSupported=true
MinTxFee=15000
BlockTime=150
FeePerByte=15
Confirmations=0
`
}

// DGB digibyte
func DGB(wallet XWallet) string {
	return `[DGB]
Title=Digibyte
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 14022
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=30
Scrwallet.IPtPrefix=5
SecretPrefix=128
COIN=100000000
MinimumAmount=0
TxVersion=1
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=false
ImportWithNoScanSupported=true
MinTxFee=100000
BlockTime=60
FeePerByte=100
Confirmations=0
`
}

// DYN dynamic
func DYN(wallet XWallet) string {
	return `[DYN]
Title=Dynamic
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 31350
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=30
Scrwallet.IPtPrefix=10
SecretPrefix=140
COIN=100000000
MinimumAmount=0
TxVersion=1
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=false
ImportWithNoScanSupported=false
MinTxFee=40000
BlockTime=128
FeePerByte=80
Confirmations=0
`
}

// DOGE dogecoin
func DOGE(wallet XWallet) string {
	return `[DOGE]
Title=Dogecoin
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 22555
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=30
Scrwallet.IPtPrefix=22
SecretPrefix=158
COIN=100000000
MinimumAmount=0
TxVersion=1
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=false
ImportWithNoScanSupported=true
MinTxFee=100000000
BlockTime=60
FeePerByte=100000
Confirmations=0
`
}

// PIVX dogecoin
func PIVX(wallet XWallet) string {
	return `[PIVX]
Title=Pivx
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 51473
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=30
Scrwallet.IPtPrefix=13
SecretPrefix=212
COIN=100000000
MinimumAmount=0
TxVersion=1
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=false
ImportWithNoScanSupported=true
MinTxFee=100000
BlockTime=60
FeePerByte=110
Confirmations=0
`
}

// VIA viacoin
func VIA(wallet XWallet) string {
	return `[VIA]
Title=Viacoin
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 5222
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=71
Scrwallet.IPtPrefix=33
SecretPrefix=199
COIN=100000000
MinimumAmount=0
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=false
ImportWithNoScanSupported=false
FeePerByte=110
MinTxFee=60000
TxVersion=1
BlockTime=24
Confirmations=0
`
}

// VTC vertcoin
func VTC(wallet XWallet) string {
	return `[VTC]
Title=Vertcoin
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 5888
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=71
Scrwallet.IPtPrefix=5
SecretPrefix=199
COIN=100000000
MinimumAmount=0
TxVersion=1
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=false
ImportWithNoScanSupported=false
MinTxFee=100000
BlockTime=150
FeePerByte=200
Confirmations=0
`
}

// MUE monetaryunit
func MUE(wallet XWallet) string {
	return `[MUE]
Title=MonetaryUnit
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 29683
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=16
Scrwallet.IPtPrefix=76
SecretPrefix=126
COIN=100000000
MinimumAmount=0
TxVersion=1
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=false
ImportWithNoScanSupported=true
MinTxFee=100000
BlockTime=40
FeePerByte=300
Confirmations=0
`
}

// NMC namecoin
func NMC(wallet XWallet) string {
	return `[NMC]
Title=Namecoin
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 8336
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=52
Scrwallet.IPtPrefix=13
SecretPrefix=180
COIN=100000000
MinimumAmount=0
TxVersion=1
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=false
ImportWithNoScanSupported=true
MinTxFee=100000
BlockTime=600
FeePerByte=100
Confirmations=0
`
}

// QTUM qtum
func QTUM(wallet XWallet) string {
	return `[QTUM]
Title=Qtum
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 3889 testnet port
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=58
Scrwallet.IPtPrefix=50
SecretPrefix=128
COIN=100000000
MinimumAmount=0
TxVersion=1
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=false
ImportWithNoScanSupported=true
MinTxFee=20000
BlockTime=150
FeePerByte=20
Confirmations=0
`
}

// LBC lbry credits
func LBC(wallet XWallet) string {
	return `[LBC]
Title=LBRY Credits
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 9245
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=85
Scrwallet.IPtPrefix=122
SecretPrefix=28
COIN=100000000
MinimumAmount=0
TxVersion=1
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=false
ImportWithNoScanSupported=true
MinTxFee=200000
BlockTime=150
FeePerByte=200
Confirmations=0
`
}

// MONA monacoin
func MONA(wallet XWallet) string {
	return `[MONA]
Title=Monacoin
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 9402
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=50
Scrwallet.IPtPrefix=55
SecretPrefix=176
COIN=100000000
MinimumAmount=0
TxVersion=1
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=false
ImportWithNoScanSupported=true
MinTxFee=200000
BlockTime=90
FeePerByte=200
Confirmations=0
`
}

// BLOCK blocknet
func BLOCK(wallet XWallet) string {
	return `[BLOCK]
Title=Blocknet
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 41414
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=26
Scrwallet.IPtPrefix=28
SecretPrefix=154
COIN=100000000
MinimumAmount=0
TxVersion=1
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=true
ImportWithNoScanSupported=true
MinTxFee=10000
BlockTime=60
FeePerByte=10
Confirmations=0
`
}

// FAIR faircoin
func FAIR(wallet XWallet) string {
	return `[FAIR]
Title=Faircoin
Address=` + wallet.Address + `
Ip=` + wallet.IP + ` # 127.0.0.1
Port=` + wallet.RPCPort + ` # 40405
Username=` + wallet.RPCUser + `
Password=` + wallet.RPCPass + `
AddressPrefix=95
Scrwallet.IPtPrefix=36
SecretPrefix=223
COIN=100000000
MinimumAmount=0
TxVersion=1
DustAmount=0
CreateTxMethod=BTC
GetNewKeySupported=true
ImportWithNoScanSupported=true
MinTxFee=30000
BlockTime=210
FeePerByte=30
Confirmations=0
`
}
