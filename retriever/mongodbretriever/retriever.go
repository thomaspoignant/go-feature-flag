package mongodbretriever

import (
	"context"
	"encoding/json"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


// Retriever is a configuration struct for a MongoDB connection and Collection.
type Retriever struct {
	Uri          string
	Collection   string
	Database     string
	dbConnection *mongo.Database
}

// Retrieve is reading flag configuration from mongodb collection and returning it
func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {

	if r.dbConnection == nil {

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(r.Uri))
		if err != nil {
			return nil, err
		}
		r.dbConnection = client.Database(r.Database)
	}
	//TODO: how to gracefully disconnect?
	// defer func () {
	// 	if err := client.Disconnect(ctx); err != nil {
	// 		panic(err)
	// 	}
	// }()

	opt := options.CollectionOptions{}
	opt.SetBSONOptions(&options.BSONOptions{OmitZeroStruct: true})

	coll := r.dbConnection.Collection(r.Collection, &opt)

	cursor, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	ffDocs := make(map[string]bson.M)

	for cursor.Next(ctx) {
		var doc bson.M

		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		if val, ok := doc["flag"]; ok {
			delete(doc, "flag")
			if str, ok := val.(string); ok {
				ffDocs[str] = doc
			} else {
				return nil, errors.New("flag key does not have a string as value!")
			}
		} else {
			return nil, errors.New("No 'flag' entry found")
		}
	}

	flags, err := json.Marshal(ffDocs)
	if err != nil {
		return nil, err
	}

	return flags, nil
}
