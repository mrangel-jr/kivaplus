package app

import (
	"kivaplus/backend/lambda/api"
	"kivaplus/backend/lambda/database"
)

type App struct {
	ApiHandler api.ApiHandler
}

func NewApp() App {
	//init dbStore
	db := database.NewDynamoDBClient()
	apiHandler := api.NewApiHandler(db)

	return App{
		ApiHandler: apiHandler,
	}
}
