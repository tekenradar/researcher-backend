package db

import (
	"errors"
)

// TODO: create notification object for study key and topic (contacts)
// TODO: update notification object for study key and topic with new email
// TODO: update notification object for study key and topic with removing email
// TODO: fetch notification object for study key and topic

func (dbService *ResearcherDBService) DeleteEmailAllNotificationsForStudy(studyKey string) (err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	if studyKey == "" {
		return errors.New("studyKey must be defined")
	}

	err = dbService.collectionRefEmailNotifications(studyKey).Drop(ctx)
	return
}
