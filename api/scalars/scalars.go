package scalars

import (
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const timeFormat = time.RFC3339

var MongoDate = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "MongoDate",
	Description: "The `MongoDate` scalar represents a DateTime.",
	Serialize: func(value interface{}) interface{} {
		switch value := value.(type) {
		case primitive.DateTime:
			return time.Unix(0, int64(value)).Format(timeFormat)
		case *primitive.DateTime:
			return time.Unix(0, int64(*value)).Format(timeFormat)
		default:
			return nil
		}
	},
	ParseValue: func(value interface{}) interface{} {
		switch value := value.(type) {
		case string:
			d, _ := time.Parse(timeFormat, value)
			return d
		case *string:
			d, _ := time.Parse(timeFormat, *value)
			return d
		default:
			return nil
		}
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		switch valueAST := valueAST.(type) {
		case *ast.StringValue:
			d, _ := time.Parse(timeFormat, valueAST.Value)
			return d
		default:
			return nil
		}
	},
})

// https://www.bradcypert.com/using-mongos-objectids-with-go-graphql/
var ObjectID = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "BSON",
	Description: "The `bson` scalar type represents a BSON Object.",
	// Serialize serializes `bson.ObjectId` to string.
	Serialize: func(value interface{}) interface{} {
		switch value := value.(type) {
		case primitive.ObjectID:
			return value.Hex()
		case *primitive.ObjectID:
			v := *value
			return v.Hex()
		default:
			return nil
		}
	},
	// ParseValue parses GraphQL variables from `string` to `bson.ObjectId`.
	ParseValue: func(value interface{}) interface{} {
		switch value := value.(type) {
		case string:
			id, _ := primitive.ObjectIDFromHex(value)
			return id
		case *string:
			id, _ := primitive.ObjectIDFromHex(*value)
			return id
		default:
			return nil
		}
		return nil
	},
	// ParseLiteral parses GraphQL AST to `bson.ObjectId`.
	ParseLiteral: func(valueAST ast.Value) interface{} {
		switch valueAST := valueAST.(type) {
		case *ast.StringValue:
			id, _ := primitive.ObjectIDFromHex(valueAST.Value)
			return id
		}
		return nil
	},
})
