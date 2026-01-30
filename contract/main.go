package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	_ "vsc-function-calls/sdk" // ensure sdk is imported

	"vsc-function-calls/sdk"
)

func main() {

}

//go:wasmexport c_log
func LogSometing(a *string) *string {
	sdk.Log(*a)
	return a
}

// payload: keyname|value
//
//go:wasmexport c_state_set
func MyStateSet(payload *string) *string {
	in := *payload
	key := nextField(&in)
	value := nextField(&in)
	require(in == "", "too many arguments")
	sdk.StateSetObject(key, value)
	return &key
}

//go:wasmexport c_state_get
func MyStateGet(key *string) *string {
	value := sdk.StateGetObject(*key)
	return value
}

//go:wasmexport c_state_del
func MyStateDelete(key *string) *string {
	sdk.StateDeleteObject(*key)
	return nil
}

//go:wasmexport c_get_env
func MyGetEnv(_ *string) *string {
	envs := sdk.GetEnv()
	// Marshal to JSON
	jsonBytes, err := json.Marshal(envs)
	if err != nil {
		sdk.Abort("failed to convert env to json")
	}

	// Convert to string
	jsonString := string(jsonBytes)
	return &jsonString
}

//go:wasmexport c_get_env_string
func MyGetEnvString(_ *string) *string {
	envs := sdk.GetEnvStr()
	return &envs
}

//go:wasmexport c_get_envkey
func MyGetEnvKey(a *string) *string {
	envVal := sdk.GetEnvKey(*a)
	return envVal
}

//go:wasmexport c_get_balance
func MyGetBalance(a *string) *string {
	bal := sdk.GetBalance(sdk.Address(*a), sdk.AssetHive) // result in terms of mHIVE/mHBD
	balStr := strconv.FormatUint(uint64(bal), 10)
	return &balStr
}

//go:wasmexport c_hive_draw
func MyHiveDraw(_ *string) *string {
	sdk.HiveDraw(1, sdk.AssetHive) // Draw 0.001 HIVE from caller
	return nil
}

//go:wasmexport c_hive_transfer
func MyHiveTransfer(receiver *string) *string {
	sdk.HiveTransfer(sdk.Address(*receiver), 1, sdk.AssetHive) // Transfer 0.001 HIVE from contract
	return nil
}

//go:wasmexport c_hive_withdraw
func MyHiveWithdraw(receiver *string) *string {
	sdk.HiveWithdraw(sdk.Address(*receiver), 1, sdk.AssetHive) // Withdraw 0.001 HIVE from contract
	return nil
}

// payload: contractId|key
//
//go:wasmexport c_contract_read
func MyContractStateGet(payload *string) *string {
	in := *payload
	contractId := nextField(&in)
	key := nextField(&in)
	require(in == "", "too many arguments")
	return sdk.ContractStateGet(contractId, key)
}

// payload: contractId|method|payload
// intent is executed to contract and then "copied" to other contract call

//go:wasmexport c_contract_call
func MyContractCall(payload *string) *string {
	in := *payload
	contractId := nextField(&in)
	method := nextField(&in)
	callPayload := in // all the rest is the payload
	callOptions := &sdk.ContractCallOptions{
		Intents: nil,
	}
	if ta := GetFirstTransferAllow(sdk.GetEnv().Intents); ta != nil {
		amt := uint64(ta.Limit * 1000)
		sdk.HiveDraw(int64(amt), ta.Token)
		callOptions.Intents = []sdk.Intent{{
			Type: "transfer.allow",
			Args: map[string]string{
				"token": ta.Token.String(),
				"limit": fmt.Sprintf("%.3f", ta.Limit),
			},
		}}
	}

	// Call another contract with same allowance as initial tx intent
	ret := sdk.ContractCall(contractId, method, callPayload, callOptions)
	return ret
}

// -- helpers --
func require(cond bool, msg string) {
	if !cond {
		sdk.Abort(msg)
	}
}

func nextField(s *string) string {
	i := strings.IndexByte(*s, '|')
	if i < 0 {
		f := *s
		*s = ""
		return f
	}
	f := (*s)[:i]
	*s = (*s)[i+1:]
	return f
}

type TransferAllow struct {
	Limit float64
	Token sdk.Asset
}

var validAssets = []string{sdk.AssetHbd.String(), sdk.AssetHive.String()}

func isValidAsset(token string) bool {
	for _, a := range validAssets {
		if token == a {
			return true
		}
	}
	return false
}

func GetFirstTransferAllow(intents []sdk.Intent) *TransferAllow {
	for _, intent := range intents {
		if intent.Type == "transfer.allow" {
			token := intent.Args["token"]
			if !isValidAsset(token) {
				sdk.Abort("invalid intent token")
			}
			limitStr := intent.Args["limit"]
			limit, err := strconv.ParseFloat(limitStr, 64)
			if err != nil {
				sdk.Abort("invalid intent limit")
			}
			return &TransferAllow{
				Limit: limit,
				Token: sdk.Asset(token),
			}
		}
	}
	return nil
}
