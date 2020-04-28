package gCloud

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/logging"
	firebase "firebase.google.com/go"
)

// LogMessage is a simple struct to ensure JSON formatting in logs
type LogMessage struct {
	Message string
}

// CloudLogger sets up a connection to Google Cloud Logging for the funciton
func CloudLogger(ctx context.Context, projectID, logName string) (*logging.Logger, error) {
	client, err := logging.NewClient(ctx, projectID)
	defer client.Close()
	return client.Logger(logName), err
}

// FireStoreCollection sets up a connetion to Firebase and fetches a connection to the desired FireStore collection
func FireStoreCollection(ctx context.Context, databaseCollection, firebaseDomain, projectID string, lg *logging.Logger) (*firestore.CollectionRef, error) {
	conf := &firebase.Config{DatabaseURL: fmt.Sprintf("https://%s.%s", projectID, firebaseDomain)}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error initializing Firebase app: %s", err)}})
		return nil, err
	}
	fsClient, err := app.Firestore(ctx)
	if err != nil {
		lg.Log(logging.Entry{Severity: logging.Error, Payload: LogMessage{Message: fmt.Sprintf("error initializing FireStore client: %s", err)}})
		return nil, err
	}
	return fsClient.Collection(databaseCollection), nil
}
