package controllers

import (
	"context"
	"log"
	"time"

	svc "cloudminds.com/harix/cc-server/services"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Connect - finding a running MongoDB Instance
func (s *CCServer) Connect() {

	// DB Config
	clientOptions := options.Client().ApplyURI(s.Config.MongoServerURI)
	client, err := mongo.NewClient(clientOptions)

	// Setup a context required by mongo.Connect

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	err = client.Connect(ctx)

	//Cancel context to avoid memory leak

	defer cancel()

	// Ping our db connection

	err = client.Ping(context.Background(), readpref.Primary())

	if err != nil {

		log.Fatal("Couldn't connect to the database", err)

	} else {
		log.Println("Connected!")
	}

	db := client.Database("go_mongo")
	svc.FamilyCollection(db)
	svc.CCRecordCollection(db)
	svc.InstCollection(db)
	svc.AdminCollection(db)
	svc.RegCodeCollection(db)
	svc.SurveyCollection(db)

	return
}
