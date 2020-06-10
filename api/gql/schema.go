package gql

import (
	"context"
	"log"
	"time"

	"github.com/graphql-go/graphql"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/kajchang/ourcampaigns-api/api/scalars"
)

const documentLimit = 1000

var limitArg = &graphql.ArgumentConfig{
	Type:         graphql.Int,
	DefaultValue: documentLimit,
	Description:  "Maximum number of documents to return",
}

var skipArg = &graphql.ArgumentConfig{
	Type:         graphql.Int,
	DefaultValue: 0,
	Description:  "Number of documents to skip",
}

func BuildGraphQLSchema() graphql.Schema {
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
		filterArgFields := make(graphql.InputObjectConfigFieldMap)
		sortArgFields := make(graphql.InputObjectConfigFieldMap)
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
				fieldType = scalars.ObjectID
			}
			typeFields[k] = &graphql.Field{
				Type: fieldType,
			}
			filterArgFields[k] = &graphql.InputObjectFieldConfig{
				Type: fieldType,
			}
			sortArgFields[k] = &graphql.InputObjectFieldConfig{
				Type: graphql.Int,
			}
		}

		queryFields[col.Name()] = &graphql.Field{
			Type: graphql.NewList(graphql.NewObject(
				graphql.ObjectConfig{
					Name:   col.Name(),
					Fields: typeFields,
				},
			)),
			Args: graphql.FieldConfigArgument{
				"filter": &graphql.ArgumentConfig{
					Type: graphql.NewInputObject(
						graphql.InputObjectConfig{
							Name:   col.Name() + "FilterArgs",
							Fields: filterArgFields,
						},
					),
					DefaultValue: bson.M(nil),
				},
				"sort": &graphql.ArgumentConfig{
					Type: graphql.NewInputObject(
						graphql.InputObjectConfig{
							Name:   col.Name() + "SortArgs",
							Fields: sortArgFields,
						},
					),
					DefaultValue: bson.M(nil),
				},
				"limit": limitArg,
				"skip":  skipArg,
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()

				limit := int64(params.Args["limit"].(int))
				if limit > documentLimit {
					limit = documentLimit
				}
				skip := int64(params.Args["skip"].(int))

				filter := params.Args["filter"]
				sort := params.Args["sort"]

				cur, err := col.Find(ctx, filter, &options.FindOptions{Sort: sort, Skip: &skip, Limit: &limit})
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
