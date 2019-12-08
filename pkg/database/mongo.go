package database

import (
	"context"
	"log"
	"time"

	"github.com/gusmin/gate/pkg/backend"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/mgo.v2/bson"
)

type Repository interface {
	Contains(ctx context.Context, machines []backend.Machine) bool
}

type MongoRepository struct {
	dbURI  string
	dbName string
	db     *mongo.Database
}

func NewMongoRepository(uri, name string) *MongoRepository {
	return &MongoRepository{
		dbURI:  uri,
		dbName: name,
	}
}

func (r *MongoRepository) Connect(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(r.dbURI))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	r.db = client.Database(r.dbName)
}

func (r *MongoRepository) Contain(ctx context.Context, machines []backend.Machine) bool {
	collection := r.db.Collection("machines")
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	res, err := collection.Find(ctx, bson.M{"id": bson.ObjectIdHex()})
}
