package db

import (
	"errors"

	"github.com/tekenradar/researcher-backend/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *ResearcherDBService) AddNotificationSubscription(substudyKey string, sub types.NotificationSubscription) (string, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	res, err := dbService.collectionRefEmailNotifications(substudyKey).InsertOne(ctx, sub)
	if err != nil {
		return "", err
	}
	id := res.InsertedID.(primitive.ObjectID)
	return id.Hex(), err
}

func (dbService *ResearcherDBService) FindNotificationSubscriptions(substudyKey string, topic string) (subs []types.NotificationSubscription, err error) {
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
	cur, err := dbService.collectionRefEmailNotifications(substudyKey).Find(ctx, filter, &opts)
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

func (dbService *ResearcherDBService) DeleteNotificationSubscription(substudyKey string, id string) (count int64, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if substudyKey == "" {
		return 0, errors.New("substudyKey must be defined")
	}

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	res, err := dbService.collectionRefEmailNotifications(substudyKey).DeleteOne(ctx, filter)
	return res.DeletedCount, err
}

func (dbService *ResearcherDBService) DeleteEmailAllNotificationsForStudy(substudyKey string) (err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	if substudyKey == "" {
		return errors.New("substudyKey must be defined")
	}

	err = dbService.collectionRefEmailNotifications(substudyKey).Drop(ctx)
	return
}
