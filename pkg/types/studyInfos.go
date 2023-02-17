package types

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StudyInfo struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key           string             `bson:"key" json:"key"`
	Name          string             `bson:"name" json:"name"`
	Description   string             `bson:"description" json:"description"`
	StudyColor    string             `bson:"studyColor" json:"studyColor"`
	AccessControl struct {
		Emails []string `bson:"emails" json:"emails"`
	} `bson:"accessControl" json:"accessControl"`
	Features struct {
		DatasetExporter bool `bson:"datasetExporter" json:"datasetExporter"`
		Contacts        bool `bson:"contacts" json:"contacts"`
	} `bson:"features" json:"features"`
	AvailableDatasets    []DatasetInfo `bson:"availableDatasets" json:"availableDatasets"`
	ContactFeatureConfig struct {
		IncludeWithParticipantFlags map[string]string `bson:"includeWithParticipantFlags" json:"includeWithParticipantFlags"`
	} `bson:"contactFeatureConfig" json:"contactFeatureConfig"`
}

type DatasetInfo struct {
	ID             string   `bson:"id" json:"id"`
	SurveyKey      string   `bson:"surveyKey" json:"surveyKey"`
	Name           string   `bson:"name" json:"name"`
	ExcludeColumns []string `bson:"excludeColumns" json:"excludeColumns"`
	StartDate      int64    `bson:"startDate" json:"startDate"`
	EndDate        int64    `bson:"endDate" json:"endDate"`
}
