package main

import (
	"fmt"
	"log"
	"os"

	"github.com/friendsofgo/graphiql"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"

	"github.com/kajchang/ourcampaigns-api/api/gql"
)

type query struct {
	Query string `json:"query"`
}

var schema graphql.Schema = gql.BuildGraphQLSchema()

func handleGraphQLRequest(c *gin.Context) {
	var q query
	err := c.BindJSON(&q)
	if err != nil {
		c.JSON(400, map[string]string{
			"error": err.Error(),
		})
		return
	}

	res := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: q.Query,
	})

	status := 200
	if len(res.Errors) > 0 {
		status = 500
	}
	c.JSON(status, res)
}

func handleGraphiQLRequest(c *gin.Context) {
	graphiqlHandler, err := graphiql.NewGraphiqlHandler("/")
	if err != nil {
		log.Fatalf("error setting up graphiql: %s", err)
	}

	graphiqlHandler.ServeHTTP(c.Writer, c.Request)
}

func main() {
	_, ok := os.LookupEnv("HEROKU")
	if ok {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.POST("/", handleGraphQLRequest)
	r.GET("/graphiql", handleGraphiQLRequest)
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}
	r.Run(fmt.Sprintf(":%s", port))
}