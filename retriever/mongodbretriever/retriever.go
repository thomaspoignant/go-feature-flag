package mongodbretriever

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/thomaspoignant/go-feature-flag/retriever"
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
	dbClient     *mongo.Client
	status       string
}

// type InitializableRetriever interface {
// 	Retrieve(ctx context.Context) ([]byte, error)
// 	Init(ctx context.Context) error
// 	Shutdown(ctx context.Context) error
// 	Status() Status
// }

func (r *Retriever) Init(ctx context.Context) error {
	if r.dbConnection == nil {
		r.status = retriever.RetrieverNotReady

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(r.Uri))
		if err != nil {
			r.status = retriever.RetrieverError
			return err
		}
		r.dbClient = client
		r.dbConnection = client.Database(r.Database)
		r.status = retriever.RetrieverReady
	}
	return nil
}

func (r *Retriever) Status(ctx context.Context) retriever.Status {
	return r.status
}

func (r *Retriever) Shutdown(ctx context.Context) error {
	return r.dbClient.Disconnect(ctx)
}

// Retrieve is reading flag configuration from mongodb collection and returning it
func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {

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
