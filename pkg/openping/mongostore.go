package ping

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoStore contains information required to establish a mondodb connection and read/write data
type MongoStore struct {
	MongoDBURL string
	Client     *mongo.Client
}

// NewMongoStore is a factory for MongoDB backed Document Storage
// The default attempts to connect to mongodb
func NewMongoStore(mongoDBURL string, user string, password string) (*MongoStore, error) {
	mongoURI := fmt.Sprintf("mongodb://%v:27017", mongoDBURL)
	if user != "" && password != "" {
		mongoURI = fmt.Sprintf("mongodb://%s:%s@%v:27017", mongoDBURL, user, password)
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	// Ping the new mongo server to see if we have a successful connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Printf("Error pinging configured mongoURL: %v", err.Error())
		return nil, err
	}

	return &MongoStore{
		MongoDBURL: mongoDBURL,
		Client:     client,
	}, nil
}

// Update stores a document in a Mongo Document Store
// It iterates through the 4 data models we have,
// 		Uptime,
// 		Latency,
// 		Metadata,
//		ContentSizes (not yet implemented)
func (ms *MongoStore) Update(uptime Uptime, latency Latency, meta Metadata, size ContentSizes) (err error) {
	uptimeCollection := ms.Client.Database("openping").Collection("uptime")
	uctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	uptimePayload := bson.M{
		"up":        uptime.Up,
		"timestamp": uptime.Timestamp,
		"rc":        uptime.RC,
		"url":       uptime.URL,
		"locale":    uptime.Locale,
		"country":   uptime.Country,
	}
	_, err = uptimeCollection.InsertOne(uctx, uptimePayload)
	if err != nil {
		return err
	}

	latencyCollection := ms.Client.Database("openping").Collection("latency")
	lctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	latencyPayload := bson.M{
		"dnslookup":    latency.DNSLookup,
		"tlshandshake": latency.TLSHandshake,
		"ttfb":         latency.TTFB,
		"total":        latency.TotalLatency,
		"timestamp":    latency.Timestamp,
		"url":          latency.URL,
	}
	_, err = latencyCollection.InsertOne(lctx, latencyPayload)
	if err != nil {
		return err
	}

	metadataCollection := ms.Client.Database("openping").Collection("metadata")
	mctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	metadataPayload := bson.M{
		"bytes":     meta.Bytes,
		"document":  meta.Document,
		"sha256sum": meta.SHASum,
		"timestamp": meta.Timestamp,
		"url":       meta.URL,
	}
	_, err = metadataCollection.InsertOne(mctx, metadataPayload)
	if err != nil {
		return err
	}

	return nil
}
