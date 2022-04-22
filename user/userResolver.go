package user

import (
	"github.com/graphql-go/graphql"
)

var CreateUser = &graphql.Field{
	Type:        MutationReturnType, // return type for this field
	Description: "Create a new user",
	Args: graphql.FieldConfigArgument{
		"name": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"email": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"password": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
	Resolve: CreateUserService,
}
