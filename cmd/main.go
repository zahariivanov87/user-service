package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"time"

	"com.user.com/user/internal/notifier"
	"com.user.com/user/internal/user"
	"com.user.com/user/internal/user/store"
	"com.user.com/user/internal/userview"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gocloud.dev/gcp"
	"gocloud.dev/pubsub/gcppubsub"

	_ "github.com/lib/pq"
)

func main() {
	// Create subscription to Pub/Sub topic in order to send notifications
	enableSubscription := os.Getenv("ENABLE_GCP_SUBSCRIPTION")
	var (
		pubsubNotifier    *notifier.PubSubNotifier
		shouldUseNotifier bool
	)

	if enableSubscription != "" && enableSubscription == "true" {
		ctx := context.Background()
		// Your GCP credentials.
		// See https://cloud.google.com/docs/authentication/production
		// for more info on alternatives.
		creds, err := gcp.DefaultCredentials(ctx)
		if err != nil {
			panic(err)
		}
		// Open a gRPC connection to the GCP Pub/Sub API.
		conn, cleanup, err := gcppubsub.Dial(ctx, creds.TokenSource)
		if err != nil {
			panic(err)
		}
		defer cleanup()

		// Construct a PublisherClient using the connection.
		pubClient, err := gcppubsub.PublisherClient(ctx, conn)
		if err != nil {
			panic(err)
		}
		defer pubClient.Close()

		// Construct a *pubsub.Topic.
		topic, err := gcppubsub.OpenTopicByPath(pubClient, "projects/myprojectID/topics/example-topic", nil)
		if err != nil {
			panic(err)
		}
		defer topic.Shutdown(ctx)
		topicURL := os.Getenv("GCP_USER_TOPIC_URL")
		if topicURL == "" {
			panic(errors.New("'GCP_USER_TOPIC_URL' must be provided"))
		}
		// Construct a *pubsub.Topic.
		topic, err = gcppubsub.OpenTopicByPath(pubClient, topicURL, nil)
		if err != nil {
			panic(err)
		}
		defer topic.Shutdown(ctx)
		pubsubNotifier = notifier.NewPubSubNotifier(*topic)
		shouldUseNotifier = true
	}

	db := openConnection("postgres", "postgresql://root@cockroachdb:26257/defaultdb?sslmode=disable", 16, 8, time.Second*time.Duration(300))
	// Create instance of User store
	userStore := store.NewStore(db)
	userStore.InitDBTables(context.Background())

	// Create instance of User manager
	userManager := user.NewManager(userStore, pubsubNotifier, shouldUseNotifier)

	// Create user endpoint
	createUserEndpoint := userview.NewCreateUserEndpoint(userManager)
	// Get All Users endpoint
	getAllUsersEndpoint := userview.NewGetAllUsersEndpoint(userManager)
	//Modify user endpoint
	modifyUserEndpoint := userview.NewUpdateUserEndpoint(userManager)
	// Delete user endpoint
	deleteUserEndpoint := userview.NewDeleteUserEndpoint(userManager)

	// Create router and bind user handlers
	router := mux.NewRouter()
	router.HandleFunc("/api/public/v1/users", createUserEndpoint.ServeHTTP).Methods(http.MethodPost)
	router.HandleFunc("/api/public/v1/users", getAllUsersEndpoint.ServeHTTP).Methods(http.MethodGet)
	router.HandleFunc("/api/public/v1/users/{userID}", deleteUserEndpoint.ServeHTTP).Methods(http.MethodDelete)
	router.HandleFunc("/api/public/v1/users/{userID}", modifyUserEndpoint.ServeHTTP).Methods(http.MethodPut)

	logrus.Info("starting web server")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		logrus.WithError(err).Error("ListenAndServe exited with error")
	}

}

func openConnection(
	driverName,
	connectionStr string,
	maxOpenDBConnections, maxIdleDBConnections int,
	maxLifetimeDBConnections time.Duration,
) sql.DB {

	db, err := sql.Open(driverName, connectionStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"driver":     driverName,
			"connection": connectionStr,
		}).WithError(err).Fatalln("failed to connect to the database")
		panic(err)
	}
	db.SetMaxOpenConns(maxOpenDBConnections)
	db.SetMaxIdleConns(maxIdleDBConnections)
	db.SetConnMaxLifetime(maxLifetimeDBConnections)

	return *db
}

// func createPublisherClient(ctx context.Context) pubsub.PublisherClient {
// 	// Your GCP credentials.
// 	// See https://cloud.google.com/docs/authentication/production
// 	// for more info on alternatives.
// 	creds, err := gcp.DefaultCredentials(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	// Open a gRPC connection to the GCP Pub/Sub API.
// 	conn, cleanup, err := gcppubsub.Dial(ctx, creds.TokenSource)
// 	if err != nil {
// 		return err
// 	}
// 	defer cleanup()

// 	// Construct a PublisherClient using the connection.
// 	pubClient, err := gcppubsub.PublisherClient(ctx, conn)
// 	if err != nil {
// 		return err
// 	}
// 	return pubClient
// }
