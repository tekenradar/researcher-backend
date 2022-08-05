package db

import (
	"github.com/tekenradar/researcher-backend/pkg/types"
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
// TODO: remove entry / where tit's not marked as permanent
