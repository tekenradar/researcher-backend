package db

import (
	"github.com/tekenradar/researcher-backend/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *ResearcherDBService) SaveStudyInfo(studyInfo types.StudyInfo) (types.StudyInfo, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"key": studyInfo.Key}

	upsert := true
	rd := options.After
	options := options.FindOneAndReplaceOptions{
		Upsert:         &upsert,
		ReturnDocument: &rd,
	}
	elem := types.StudyInfo{}
	err := dbService.collectionRefStudyInfos().FindOneAndReplace(
		ctx, filter, studyInfo, &options,
	).Decode(&elem)
	return elem, err
}

// TODO:
// delete study info
// find study infos (list of keys)

func (dbService *ResearcherDBService) FindStudyInfosByKeys(studyKeys []string) (studyInfos []types.StudyInfo, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"studyKey": bson.M{
			"$in": studyKeys,
		},
	}

	batchSize := int32(32)
	opts := options.FindOptions{
		BatchSize: &batchSize,
	}
	cur, err := dbService.collectionRefStudyInfos().Find(ctx, filter, &opts)
	if err != nil {
		return studyInfos, err
	}
	defer cur.Close(ctx)

	studyInfos = []types.StudyInfo{}
	for cur.Next(ctx) {
		var result types.StudyInfo
		err := cur.Decode(&result)

		if err != nil {
			return studyInfos, err
		}

		studyInfos = append(studyInfos, result)
	}
	if err := cur.Err(); err != nil {
		return studyInfos, err
	}

	return studyInfos, nil
}

func (dbService *ResearcherDBService) FindAllStudyInfos() (studyInfos []types.StudyInfo, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	batchSize := int32(32)
	opts := options.FindOptions{
		BatchSize: &batchSize,
	}
	cur, err := dbService.collectionRefStudyInfos().Find(ctx, filter, &opts)
	if err != nil {
		return studyInfos, err
	}
	defer cur.Close(ctx)

	studyInfos = []types.StudyInfo{}
	for cur.Next(ctx) {
		var result types.StudyInfo
		err := cur.Decode(&result)

		if err != nil {
			return studyInfos, err
		}

		studyInfos = append(studyInfos, result)
	}
	if err := cur.Err(); err != nil {
		return studyInfos, err
	}

	return studyInfos, nil
}
