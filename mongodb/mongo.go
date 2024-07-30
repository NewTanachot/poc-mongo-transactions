package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

var (
	once                    sync.Once
	client                  *MongoDb
	replicaConnectionString = "mongodb://admin:password@localhost:27017,localhost:27018,localhost:27019/?replicaSet=rs0"
	mongoDbName             = "test_database"
)

type MongoDb struct {
	Client           *mongo.Client
	DbClient         *mongo.Database
	ConnectionString string
}

func NewMongoDbClient(connectionString ...string) *MongoDb {
	once.Do(func() {
		var cStr string

		if len(connectionString) == 0 {
			cStr = replicaConnectionString
		} else {
			cStr = connectionString[0]
		}

		// Use the SetServerAPIOptions() method to set the Stable API version to 1
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.
			Client().
			ApplyURI(cStr).
			SetServerAPIOptions(serverAPI)

		// go driver use context to set timeout for this task only
		ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancle()

		// Create a new client and connect to the server
		mClient, err := mongo.Connect(ctx, opts)

		if err != nil {
			fmt.Println(err.Error())
		}

		client = &MongoDb{
			Client:           mClient,
			DbClient:         mClient.Database(mongoDbName),
			ConnectionString: cStr,
		}
	})

	if isSuccess := client.PingMongoDb(); !isSuccess {
		fmt.Println("mongo ping fail")
	}

	return client
}

func (m *MongoDb) CloseMongoDb() {
	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()

	m.Client.Disconnect(ctx)
}

func (m *MongoDb) PingMongoDb() bool {
	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()

	if err := m.Client.Ping(ctx, nil); err != nil {
		return false
	}

	return true
}

func (m *MongoDb) Execute(ctx context.Context, transactions ...func(mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	opts := options.
		Transaction().
		SetWriteConcern(writeconcern.Majority()).
		SetReadConcern(readconcern.Majority())

	session, err := m.DbClient.Client().StartSession()

	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	return session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		for _, transaction := range transactions {
			if result, err := transaction(sessCtx); err != nil {
				return result, err
			}
		}
		return nil, nil
	}, opts)
}
