package backend

import (
	gql "github.com/gusmin/graphql"
)

const addMachineLogMutation = `
	mutation addMachineLog($machineLogs: [MachineLogInput!]!) {
  	addMachineLog(machineLogs: $machineLogs) {
   	 success
  	}
	}
`

func makeAddMachineLogRequest(token string, inputs []MachineLogInput) *gql.Request {
	req := gql.NewRequest(addMachineLogMutation)
	req.Header.Set("Authorization", "JWT "+token)
	req.Var("machineLogs", inputs)
	return req
}
