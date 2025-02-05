package user

import (
	"errors"
	"fmt"
	"log"

	"github.com/graphql-go/graphql"
	"github.com/guiwoo/exercise_backend/jwtValidator"
	"github.com/guiwoo/exercise_backend/model"
	"github.com/guiwoo/exercise_backend/utils"

	"golang.org/x/crypto/bcrypt"
)

const Cost int = 10

type MutationReturn struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type LoginReturn struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
	Token string `json:"token"`
}

type FindUserReturn struct {
	Ok    bool              `json:"ok"`
	Error string            `json:"error"`
	Users []model.User_Type `json:"users"`
}

type oneUserReturn struct {
	Ok    bool            `json:"ok"`
	Error string          `json:"error"`
	User  model.User_Type `json:"user"`
}

var (
	service = model.DB_Handler()
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), Cost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

var CreateUserService = func(p graphql.ResolveParams) (interface{}, error) {
	name, _ := p.Args["name"].(string)
	email, _ := p.Args["email"].(string)
	password, _ := p.Args["password"].(string)

	var user = model.User_Type{Email: email}
	has, err := service.Get(&user)
	if err != nil {
		log.Fatal(err)
	}
	if has {
		return &MutationReturn{Ok: false, Error: "email has already taken"}, nil
	}
	password, _ = hashPassword(password)
	user.Name = name
	user.Password = password
	affected, err := service.Insert(user)
	utils.HandleErr(err)
	if affected != 1 {
		log.Fatal("Engine affected not 1 column")
		return &MutationReturn{Ok: false, Error: "Fail to create account"}, nil
	}
	return &MutationReturn{Ok: true, Error: "nil"}, nil
}

var LoginUserService = func(p graphql.ResolveParams) (interface{}, error) {
	var findUsers []model.User_Type
	email, _ := p.Args["email"].(string)
	password, _ := p.Args["password"].(string)
	err := service.In("email", email).Find(&findUsers)
	if err != nil {
		log.Fatal(err)
	}
	if len(findUsers) < 1 {
		return LoginReturn{Ok: false, Error: "could not find the email", Token: ""}, errors.New("could not find the email")
	}
	if ok := checkPasswordHash(password, findUsers[0].Password); !ok {
		return LoginReturn{Ok: false, Error: "password is not correct", Token: ""}, errors.New("password is not correct")
	}
	token := jwtValidator.GenerateToken(email, findUsers[0].ID)
	return LoginReturn{Ok: true, Error: "lala", Token: token}, nil
}

var EditUserService = func(p graphql.ResolveParams) (interface{}, error) {
	name, _ := p.Args["name"].(string)
	password, _ := p.Args["password"].(string)
	id, _ := jwtValidator.JwtValidChecker(p)
	if name != "" {
		var findUsers []model.User_Type
		err := service.In("name", name).Find(&findUsers)
		if err != nil {
			log.Fatal(err)
		}
		// duplicate check
		if len(findUsers) < 1 {
			user := new(model.User_Type)
			user.Name = name
			affected, err := service.ID(id).Update(user)
			if err != nil {
				log.Fatal(err)
			}
			if affected < 1 {
				return &MutationReturn{Ok: false, Error: "Update failed"}, nil
			}
		} else {
			return &MutationReturn{Ok: false, Error: "Duplicate Nicname"}, nil
		}
	}
	// if password is not empty
	if password != "" {
		if len(password) < 3 {
			return &MutationReturn{Ok: false, Error: "Password should be longer than 3 characters"}, nil
		}
		user := new(model.User_Type)
		newPassword, _ := hashPassword(password)
		user.Password = newPassword
		affected, err := service.ID(id).Update(user)
		if err != nil {
			log.Fatal(err)
		}
		if affected < 1 {
			return &MutationReturn{Ok: false, Error: "Update failed"}, nil
		}
	}
	return &MutationReturn{Ok: true, Error: "nil"}, nil
}

// find a user by name, don't need to log-in
var SearchUserService = func(p graphql.ResolveParams) (interface{}, error) {
	var result []model.User_Type
	name := p.Args["name"].(string)
	if len(name) < 3 {
		return &FindUserReturn{Ok: false, Error: "Search Name characters at least 3", Users: result}, nil
	}
	sql := `SELECT * from User__Type WHERE (lower(name) LIKE CONCAT('%',$1::text,'%'))`
	err := service.SQL(sql, name).Find(&result) // => 리턴 정확하게 User_Type 으로 반환
	utils.HandleErr(err)
	return &FindUserReturn{Ok: true, Users: result}, nil
}

// find user by id ,why need ? when click profile send to sepcific page
var FindUserByIdService = func(p graphql.ResolveParams) (interface{}, error) {
	var user model.User_Type
	id := p.Args["id"].(int)
	has, err := service.ID(id).Get(&user)
	utils.HandleErr(err)
	if !has {
		return &oneUserReturn{Ok: false, Error: fmt.Sprintf("Could not find this user id %d", id)}, nil
	}
	return &oneUserReturn{Ok: true, User: user}, nil
}
