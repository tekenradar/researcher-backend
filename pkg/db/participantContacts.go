package db

import (
	"time"

	"github.com/tekenradar/researcher-backend/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (dbService *ResearcherDBService) AddParticipantContact(studyKey string, pc types.ParticipantContact) (string, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	res, err := dbService.collectionRefParticipantContacts(studyKey).InsertOne(ctx, pc)
	if err != nil {
		return "", err
	}
	id := res.InsertedID.(primitive.ObjectID)
	return id.Hex(), err
}

// TODO: add note to entry
// TODO: mark entry as permanent

// Remove entries after certain time if not marked as permanent
func (dbService *ResearcherDBService) CleanUpExpiredParticipantContacts(studyKey string, deleteAfterInDays int) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	ref := time.Now().AddDate(0, 0, -deleteAfterInDays).Unix()
	filter := bson.M{
		"$and": bson.A{
			bson.M{"addedAt": bson.M{"$lt": ref}},
			bson.M{"keepContactData": false},
		},
	}
	_, err := dbService.collectionRefParticipantContacts(studyKey).DeleteMany(ctx, filter)
	return err
}
