package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/friendsofgo/graphiql"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/kajchang/ourcampaigns-api/api/scalars"
)

type query struct {
	Query string `json:"query"`
}

func buildGraphQLSchema() graphql.Schema {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %s", err)
	}

	db := client.Database("ourcampaigns")
	colNames, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Fatalf("failed to list collections: %s", err)
	}

	queryFields := make(graphql.Fields)
	for _, colName := range colNames {
		col := db.Collection(colName)

		var sampleDoc bson.D

		res := col.FindOne(ctx, bson.M{})
		err = res.Decode(&sampleDoc)
		if err != nil {
			log.Fatalf("error finding sample doc in %s collection: %s", col.Name(), err)
		}

		typeFields := make(graphql.Fields)
		for k, v := range sampleDoc.Map() {
			var fieldType graphql.Output

			switch t := v.(type) {
			default:
				log.Fatalf("unknown type in mongodb: %s", t)
			case string:
				fieldType = graphql.String
			case int32:
				fieldType = graphql.Int
			case uint8:
				fieldType = graphql.Int
			case float64:
				fieldType = graphql.Float
			case primitive.DateTime:
				fieldType = scalars.MongoDate
			case primitive.ObjectID:
				fieldType = graphql.ID
			}
			typeFields[k] = &graphql.Field{
				Type: fieldType,
			}
		}

		queryFields[col.Name()] = &graphql.Field{
			Type: graphql.NewList(graphql.NewObject(
				graphql.ObjectConfig{
					Name:   col.Name(),
					Fields: typeFields,
				},
			)),
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()

				docLimit := int64(1000)
				cur, err := col.Find(ctx, bson.M{}, &options.FindOptions{Limit: &docLimit})
				if err != nil {
					return nil, err
				}

				var res []map[string]interface{}
				for cur.Next(ctx) {
					var doc bson.D
					cur.Decode(&doc)
					res = append(res, doc.Map())
				}
				cur.Close(ctx)
				return res, nil
			},
		}
	}

	query := graphql.ObjectConfig{Name: "Query", Fields: queryFields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(query)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema; %s\n", err)
	}
	return schema
}

var schema graphql.Schema = buildGraphQLSchema()

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
	r := gin.Default()

	r.POST("/", handleGraphQLRequest)
	r.GET("/graphiql", handleGraphiQLRequest)
	r.Run()
}
