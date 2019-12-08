package backend

import (
	gql "github.com/gusmin/graphql"
)

const authQuery = `
	query auth($email: String!, $password: String!) {
		auth(email: $email, password: $password) {
			success
			token
			message
		}
	}
`

func makeAuthRequest(email, password string) *gql.Request {
	req := gql.NewRequest(authQuery)
	req.Var("email", email)
	req.Var("password", password)
	return req
}

const machinesQuery = `
	query machines {
		machines {
			id
			name
			ip
			agentPort
		}
	}
`

func makeMachinesRequest(token string) *gql.Request {
	req := gql.NewRequest(machinesQuery)
	req.Header.Set("Authorization", "JWT "+token)
	return req
}

const meQuery = `
	query userInfos {
		user: me {
			id
			email
			firstName
			lastName
			job
		}
	}
`

func makeMeRequest(token string) *gql.Request {
	req := gql.NewRequest(meQuery)
	req.Header.Set("Authorization", "JWT "+token)
	return req
}
