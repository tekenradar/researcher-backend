package db

import (
	"errors"

	"github.com/tekenradar/researcher-backend/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *ResearcherDBService) AddNotificationSubscription(studyKey string, sub types.NotificationSubscription) (string, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	res, err := dbService.collectionRefEmailNotifications(studyKey).InsertOne(ctx, sub)
	if err != nil {
		return "", err
	}
	id := res.InsertedID.(primitive.ObjectID)
	return id.Hex(), err
}

func (dbService *ResearcherDBService) FindNotificationSubscriptions(studyKey string, topic string) (subs []types.NotificationSubscription, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	if len(topic) > 0 {
		filter["topic"] = topic
	}

	batchSize := int32(32)
	opts := options.FindOptions{
		BatchSize: &batchSize,
	}
	cur, err := dbService.collectionRefEmailNotifications(studyKey).Find(ctx, filter, &opts)
	if err != nil {
		return subs, err
	}
	defer cur.Close(ctx)

	subs = []types.NotificationSubscription{}
	for cur.Next(ctx) {
		var result types.NotificationSubscription
		err := cur.Decode(&result)

		if err != nil {
			return subs, err
		}

		subs = append(subs, result)
	}
	if err := cur.Err(); err != nil {
		return subs, err
	}

	return subs, nil
}

func (dbService *ResearcherDBService) DeleteNotificationSubscription(studyKey string, id string) (count int64, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if studyKey == "" {
		return 0, errors.New("studyKey must be defined")
	}

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	res, err := dbService.collectionRefEmailNotifications(studyKey).DeleteOne(ctx, filter)
	return res.DeletedCount, err
}

func (dbService *ResearcherDBService) DeleteEmailAllNotificationsForStudy(studyKey string) (err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	if studyKey == "" {
		return errors.New("studyKey must be defined")
	}

	err = dbService.collectionRefEmailNotifications(studyKey).Drop(ctx)
	return
}
