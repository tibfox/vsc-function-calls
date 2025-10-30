package contract_test

import (
	"testing"
	"vsc-node/modules/db/vsc/contracts"
)

// admin tests
func TestBasics(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "log", []byte("my basic log"), nil, "hive:someone", true, uint(1_000_000_000), "my basic log", nil, ContractID1)
	CallContract(t, ct, "state_set", []byte("mykey|myvalue"), nil, "hive:someone", true, uint(1_000_000_000), "", nil, ContractID1)
	CallContract(t, ct, "state_get", []byte("mykey"), nil, "hive:someone", true, uint(1_000_000_000), "myvalue", nil, ContractID1)
	CallContract(t, ct, "state_del", []byte("mykey"), nil, "hive:someone", true, uint(1_000_000_000), "", nil, ContractID1)
	CallContract(t, ct, "state_get", []byte("mykey"), nil, "hive:someone", true, uint(1_000_000_000), "", nil, ContractID1)
}

func TestEnv(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "get_env", []byte(""), nil, "hive:someone", true, uint(1_000_000_000), "", nil, ContractID1)
	CallContract(t, ct, "get_env_string", []byte(""), nil, "hive:someone", true, uint(1_000_000_000), "", nil, ContractID1)
	CallContract(t, ct, "get_envkey", []byte("contract.id"), nil, "hive:someone", true, uint(1_000_000_000), "vsctestcontract1", nil, ContractID1)
}

func TestBalance(t *testing.T) {
	ct := SetupContractTest()
	CallContract(t, ct, "get_balance", []byte("hive:someone"), nil, "hive:someone", true, uint(1_000_000_000), "5000", nil, ContractID1)
	CallContract(t, ct, "hive_draw", []byte(""), nil, "hive:someone", false, uint(1_000_000_000), "", nil, ContractID1)
	CallContract(t, ct, "hive_draw", []byte(""),
		[]contracts.Intent{{Type: "transfer.allow", Args: map[string]string{"limit": "0.001", "token": "hive"}}},
		"hive:someone", true, uint(1_000_000_000), "", nil, ContractID1)
	CallContract(t, ct, "hive_transfer", []byte("hive:someoneelse"), nil, "hive:someone", true, uint(1_000_000_000), "", nil, ContractID1)
	CallContract(t, ct, "hive_draw", []byte(""),
		[]contracts.Intent{{Type: "transfer.allow", Args: map[string]string{"limit": "0.001", "token": "hive"}}},
		"hive:someone", true, uint(1_000_000_000), "", nil, ContractID1)
	CallContract(t, ct, "hive_withdraw", []byte("hive:someoneelse"), nil, "hive:someone", true, uint(1_000_000_000), "", nil, ContractID1)
}

func TestContract(t *testing.T) {
	ct := SetupContractTest()
	// store key in contract 1
	CallContract(t, ct, "state_set", []byte("mykey|myvalue"), nil, "hive:someone", true, uint(1_000_000_000), "", nil, ContractID1)
	// read key from contract 1 via contract 2
	CallContract(t, ct, "contract_read", []byte(ContractID1+"|mykey"), nil, "hive:someone", true, uint(1_000_000_000), "myvalue", nil, ContractID2)
	CallContract(t, ct, "contract_call", []byte(ContractID1+"|state_set|mykey2|myvalue2"), nil, "hive:someone", true, uint(1_000_000_000), "", nil, ContractID2)
	CallContract(t, ct, "contract_read", []byte(ContractID1+"|mykey2"), nil, "hive:someone", true, uint(1_000_000_000), "myvalue2", nil, ContractID2)

	CallContract(t, ct, "contract_call", []byte(ContractID1+"|hive_draw"),
		[]contracts.Intent{{Type: "transfer.allow", Args: map[string]string{"limit": "0.001", "token": "hive"}}},
		"hive:someone", true, uint(1_000_000_000), "", nil, ContractID2)
}
