package main

import (
	"casbin_adapter/adapter"
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/casbin/casbin/v2"
)

var (
	ctx context.Context
	db  *sql.DB
)

func main() {

	// a, _ := adapter.NewAdapter("host=localhost port=5432 user=postgres dbname=casbin_dev sslmode=disable schema='app_data'")
	a, _ := adapter.NewAdapter("host=localhost port=5432 user=postgres dbname=casbin_dev sslmode=disable")

	authEnforcer, err := casbin.NewEnforcer("conf/auth_model.conf", a)
	if err != nil {
		log.Fatal(err)
	}

	err = authEnforcer.LoadPolicy()
	if err != nil {
		log.Fatal(err)
		fmt.Println(err)
	}

	res, err := authEnforcer.AddRoleForUser("u003", "p001.READ_WRITE")
	if err != nil {
		log.Fatal(err)
		fmt.Println(err)
	}
	fmt.Println(res)

	res, err = authEnforcer.Enforce("u003", "p001", "GET")
	if err != nil {

		errorMsg := "could not find required authorization"
		fmt.Println(errorMsg)
		return
	}
	fmt.Println(res)
}
