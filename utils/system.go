package utils

import (
	"github.com/denisbrodbeck/machineid"
)

const appID = "9a621c30-566b-4231-ab79-9262810cebfc"

var MachineID = getMachineID()

func getMachineID() string {
	id, err := machineid.ProtectedID(appID)
	if err != nil {
		panic(err)
	}
	return id
}
