package mongodbretriever

import (
	"context"
	"encoding/json"

	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Retriever is a configuration struct for a MongoDB connection and Collection.
type Retriever struct {
	// MongoDB connection URI
	URI string
	// Mongodb database where flags collection is
	Database string
	// Mongodb collection where flag definitions are stored
	Collection   string
	dbConnection *mongo.Database
	dbClient     *mongo.Client
	status       string
	logger       *fflog.FFLogger
}

// Init is initializing the retriever to start fetching the flags configuration.
func (r *Retriever) Init(ctx context.Context, logger *fflog.FFLogger) error {
	r.logger = logger
	if r.dbConnection == nil {
		r.status = retriever.RetrieverNotReady

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(r.URI))
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

// Status returns the current status of the retriever
func (r *Retriever) Status() retriever.Status {
	if r == nil || r.status == "" {
		return retriever.RetrieverNotReady
	}
	return r.status
}

// Shutdown disconnects the retriever from Mongodb instance
func (r *Retriever) Shutdown(ctx context.Context) error {
	return r.dbClient.Disconnect(ctx)
}

// Retrieve Reads flag configuration from mongodb and returns it
// if a document does not comply with the specification it will be ignored
func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	opt := options.CollectionOptions{}
	opt.SetBSONOptions(&options.BSONOptions{OmitZeroStruct: true})

	coll := r.dbConnection.Collection(r.Collection, &opt)

	cursor, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

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
				r.logger.Error("flag key does not have a string as value")
			}
		} else {
			r.logger.Error("no 'flag' entry found")
		}
	}

	flags, err := json.Marshal(ffDocs)
	if err != nil {
		return nil, err
	}

	return flags, nil
}
