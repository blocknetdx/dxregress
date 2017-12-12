package chain

import "strings"

const (
	WalletBTC 	= "BTC"
	WalletLTC	= "LTC"
	WalletMONA 	= "MONA"
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
	WalletLBC	= "LBC"
	WalletFAIR	= "FAIR"
	WalletSEQ	= "SEQ"
)

func Main(wallets []string) string {
	return `[Main]
ExchangeWallets=`+strings.Join(wallets, ",")+`
FullLog=true
LogPath=/var/log/xbridge.log
ExchangeTax=300

[RPC]
Enable=false
UserName=[put]
Password=[put]
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
TxVersion=1
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
FeePerByte=200
MinTxFee=60000
TxVersion=1
BlockTime=60
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
FeePerByte=200
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
MinTxFee=10000
BlockTime=150
FeePerByte=10
Confirmations=0
`
}