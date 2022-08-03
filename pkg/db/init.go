package db

import (
	"context"
	"time"

	"github.com/coneno/logger"
	"github.com/tekenradar/researcher-backend/pkg/types"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ResearcherDBService struct {
	DBClient     *mongo.Client
	timeout      int
	DBNamePrefix string
}

func NewResearcherDBService(configs types.DBConfig) *ResearcherDBService {
	var err error
	dbClient, err := mongo.NewClient(
		options.Client().ApplyURI(configs.URI),
		options.Client().SetMaxConnIdleTime(time.Duration(configs.IdleConnTimeout)*time.Second),
		options.Client().SetMaxPoolSize(configs.MaxPoolSize),
	)
	if err != nil {
		logger.Error.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(configs.Timeout)*time.Second)
	defer cancel()

	err = dbClient.Connect(ctx)
	if err != nil {
		logger.Error.Fatal(err)
	}

	ctx, conCancel := context.WithTimeout(context.Background(), time.Duration(configs.Timeout)*time.Second)
	err = dbClient.Ping(ctx, nil)
	defer conCancel()
	if err != nil {
		logger.Error.Fatal("fail to connect to DB: " + err.Error())
	}

	ResearcherDBService := &ResearcherDBService{
		DBClient:     dbClient,
		timeout:      configs.Timeout,
		DBNamePrefix: configs.DBNamePrefix,
	}

	// TODO: create indexes

	return ResearcherDBService
}

//new Collection
func (dbService *ResearcherDBService) collectionRefStudyInfos() *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + "researcherDB").Collection("study-infos")
}

func (dbService *ResearcherDBService) collectionRefEmailNotifications(studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + "researcherDB").Collection("email-notifications" + studyKey)
}

func (dbService *ResearcherDBService) collectionRefParticipantContacts(studyKey string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.DBNamePrefix + "researcherDB").Collection("participant-contacts-" + studyKey)
}

// DB utils
func (dbService *ResearcherDBService) getContext() (ctx context.Context, cancel context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(dbService.timeout)*time.Second)
}
