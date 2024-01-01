package mongodbretriever

import (
	"context"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Targeting struct {
	query      string          `bson:"query"`
	percentage map[string]uint `bson:"percentage"`
	variation  string          `bson:"variation"`
	disable    bool            `bson:"disable"`
}

type FeatureFlag struct {
	Flag       string                 `bson:"flag"`
	Variations map[string]interface{} `bson:"variations"`
	Targeting  []Targeting            `bson:"targeting"`
}

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

	var ffDocs []bson.M

	for cursor.Next(ctx) {
		var doc bson.M

		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		ffDocs = append(ffDocs, doc)
	}

	flags, err := json.Marshal(ffDocs)
	fmt.Print(flags)
	if err != nil {
		return nil, err
	}

	return flags, nil
}
