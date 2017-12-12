package chain

import "strings"

const (
	WalletBTC 	= "BTC"
	WalletLTC	= "LTC"
	WalletSYS 	= "SYS"
	WalletDASH 	= "DASH"
	WalletDGB	= "DGB"
	WalletDYN	= "DYN"
	WalletDOGE	= "DOGE"
	WalletPIVX	= "PIVX"
	WalletVIA	= "VIA"
	WalletVTC	= "VTC"
	WalletMUE	= "MUE"
	WalletNMC	= "NMC"
	WalletQTUM	= "QTUM"
	WalletLBC	= "LBC"
	WalletMONA 	= "MONA"
	WalletBLOCK	= "BLOCK"
	WalletFAIR	= "FAIR"
)

var Wallets = map[string]string{
	WalletBTC: WalletBTC,
	WalletLTC: WalletLTC,
	WalletSYS: WalletSYS,
	WalletDASH: WalletDASH,
	WalletDGB: WalletDGB,
	WalletDYN: WalletDYN,
	WalletDOGE: WalletDOGE,
	WalletPIVX: WalletPIVX,
	WalletMONA: WalletMONA,
	WalletVIA: WalletVIA,
	WalletVTC: WalletVTC,
	WalletMUE: WalletMUE,
	WalletNMC: WalletNMC,
	WalletQTUM: WalletQTUM,
	WalletLBC: WalletLBC,
	WalletBLOCK: WalletBLOCK,
	WalletFAIR: WalletFAIR,
}

// SupportsWallet returns true if the wallet is supported.
func SupportsWallet(wallet string) bool {
	_, ok := Wallets[wallet]
	return ok
}

// DefaultXConfig returns the default xbridge config for the specified coin.
func DefaultXConfig(coin, address, ip, rpcuser, rpcpass string) string {
	switch coin {
	case WalletBTC:
		return BTC(address, ip, "8332", rpcuser, rpcpass)
	case WalletLTC:
		return LTC(address, ip, "9332", rpcuser, rpcpass)
	case WalletSYS:
		return SYS(address, ip, "8370", rpcuser, rpcpass)
	case WalletDASH:
		return DASH(address, ip, "9998", rpcuser, rpcpass)
	case WalletDGB:
		return DGB(address, ip, "14022", rpcuser, rpcpass)
	case WalletDYN:
		return DYN(address, ip, "31350", rpcuser, rpcpass)
	case WalletDOGE:
		return DOGE(address, ip, "22555", rpcuser, rpcpass)
	case WalletPIVX:
		return PIVX(address, ip, "51473", rpcuser, rpcpass)
	case WalletVIA:
		return VIA(address, ip, "5222", rpcuser, rpcpass)
	case WalletVTC:
		return VTC(address, ip, "5888", rpcuser, rpcpass)
	case WalletMUE:
		return MUE(address, ip, "29683", rpcuser, rpcpass)
	case WalletNMC:
		return NMC(address, ip, "8336", rpcuser, rpcpass)
	case WalletQTUM:
		return QTUM(address, ip, "3889", rpcuser, rpcpass)
	case WalletLBC:
		return LBC(address, ip, "9245", rpcuser, rpcpass)
	case WalletMONA:
		return MONA(address, ip, "9402", rpcuser, rpcpass)
	case WalletBLOCK:
		return BLOCK(address, ip, "41414", rpcuser, rpcpass)
	case WalletFAIR:
		return FAIR(address, ip, "40405", rpcuser, rpcpass)
	}
	return ""
}

// MAIN returns the main config section.
func MAIN(wallets []string) string {
	return `[Main]
ExchangeWallets=`+strings.Join(wallets, ",")+`
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
func BTC(address, ip, port, rpcuser, rpcpass string) string {
	return `[BTC]
Title=Bitcoin
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 8332
Username=`+rpcuser+`
Password=`+rpcpass+`
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
func LTC(address, ip, port, rpcuser, rpcpass string) string {
	return `[LTC]
Title=Litecoin
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 9332
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=48
ScriptPrefix=5
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
func SYS(address, ip, port, rpcuser, rpcpass string) string {
	return `[SYS]
Title=SysCoin2
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 8370
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=0
ScriptPrefix=5
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
func DASH(address, ip, port, rpcuser, rpcpass string) string {
	return `[DASH]
Title=Dash
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 9998
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=76
ScriptPrefix=16
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
func DGB(address, ip, port, rpcuser, rpcpass string) string {
	return `[DGB]
Title=Digibyte
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 14022
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=30
ScriptPrefix=5
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
func DYN(address, ip, port, rpcuser, rpcpass string) string {
	return `[DYN]
Title=Dynamic
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 31350
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=30
ScriptPrefix=10
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
func DOGE(address, ip, port, rpcuser, rpcpass string) string {
	return `[DOGE]
Title=Dogecoin
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 22555
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=30
ScriptPrefix=22
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
func PIVX(address, ip, port, rpcuser, rpcpass string) string {
	return `[PIVX]
Title=Pivx
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 51473
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=30
ScriptPrefix=13
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
func VIA(address, ip, port, rpcuser, rpcpass string) string {
	return `[VIA]
Title=Viacoin
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 5222
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=71
ScriptPrefix=33
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
func VTC(address, ip, port, rpcuser, rpcpass string) string {
	return `[VTC]
Title=Vertcoin
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 5888
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=71
ScriptPrefix=5
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
func MUE(address, ip, port, rpcuser, rpcpass string) string {
	return `[MUE]
Title=MonetaryUnit
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 29683
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=16
ScriptPrefix=76
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
func NMC(address, ip, port, rpcuser, rpcpass string) string {
	return `[NMC]
Title=Namecoin
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 8336
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=52
ScriptPrefix=13
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
func QTUM(address, ip, port, rpcuser, rpcpass string) string {
	return `[QTUM]
Title=Qtum
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 3889 testnet port
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=58
ScriptPrefix=50
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
func LBC(address, ip, port, rpcuser, rpcpass string) string {
	return `[LBC]
Title=LBRY Credits
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 9245
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=85
ScriptPrefix=122
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
func MONA(address, ip, port, rpcuser, rpcpass string) string {
	return `[MONA]
Title=Monacoin
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 9402
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=50
ScriptPrefix=55
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
func BLOCK(address, ip, port, rpcuser, rpcpass string) string {
	return `[BLOCK]
Title=Blocknet
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 41414
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=26
ScriptPrefix=28
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
func FAIR(address, ip, port, rpcuser, rpcpass string) string {
	return `[FAIR]
Title=Faircoin
Address=`+address+`
Ip=`+ip+` # 127.0.0.1
Port=`+port+` # 40405
Username=`+rpcuser+`
Password=`+rpcpass+`
AddressPrefix=95
ScriptPrefix=36
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