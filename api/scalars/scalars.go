package scalars

import (
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const defaultTimeFormat = time.RFC3339

var MongoDate = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "MongoDate",
	Description: "The `MongoDate` scalar represents a DateTime.",
	Serialize: func(value interface{}) interface{} {
		switch value := value.(type) {
		case primitive.DateTime:
			return time.Unix(0, int64(value)).Format(defaultTimeFormat)
		case *primitive.DateTime:
			return time.Unix(0, int64(*value)).Format(defaultTimeFormat)
		default:
			return nil
		}
	},
	ParseValue: func(value interface{}) interface{} {
		switch value := value.(type) {
		case string:
			d, _ := time.Parse(defaultTimeFormat, value)
			return d
		case *string:
			d, _ := time.Parse(defaultTimeFormat, *value)
			return d
		default:
			return nil
		}
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		switch valueAST := valueAST.(type) {
		case *ast.StringValue:
			d, _ := time.Parse(defaultTimeFormat, valueAST.Value)
			return d
		default:
			return nil
		}
	},
})
