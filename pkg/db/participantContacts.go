package db

import (
	"time"

	"github.com/tekenradar/researcher-backend/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (dbService *ResearcherDBService) UpdateKeepParticipantContactStatus(studyKey string, contactID string, value bool) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(contactID)
	filter := bson.M{"_id": _id}

	update := bson.M{"$set": bson.M{"keepContactData": value}}
	_, err := dbService.collectionRefParticipantContacts(studyKey).UpdateOne(ctx, filter, update)
	return err
}

func (dbService *ResearcherDBService) AddNoteToParticipantContact(studyKey string, contactID string, note types.ContactNote) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(contactID)
	filter := bson.M{"_id": _id}

	update := bson.M{"$push": bson.M{"notes": bson.M{
		"$each": bson.A{
			note,
		},
		"$position": 0,
	}}}
	_, err := dbService.collectionRefParticipantContacts(studyKey).UpdateOne(ctx, filter, update)
	return err
}

func (dbService *ResearcherDBService) FindParticipantContacts(studyKey string) (pcs []types.ParticipantContact, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	batchSize := int32(32)
	opts := options.FindOptions{
		BatchSize: &batchSize,
	}
	cur, err := dbService.collectionRefParticipantContacts(studyKey).Find(ctx, filter, &opts)
	if err != nil {
		return pcs, err
	}
	defer cur.Close(ctx)

	pcs = []types.ParticipantContact{}
	for cur.Next(ctx) {
		var result types.ParticipantContact
		err := cur.Decode(&result)

		if err != nil {
			return pcs, err
		}

		pcs = append(pcs, result)
	}
	if err := cur.Err(); err != nil {
		return pcs, err
	}

	return pcs, nil
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
