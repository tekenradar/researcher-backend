package runner

import (
	"time"

	"github.com/coneno/logger"
	"github.com/tekenradar/researcher-backend/pkg/db"
)

const (
	deleteAfterInDays = 7 * 12
)

type Runner struct {
	researcherDB       *db.ResearcherDBService
	timerEventCooldown int64 // how often the timer event should be performed
}

func NewRunner(researcherDB *db.ResearcherDBService, timerEventCooldown int64) *Runner {
	return &Runner{
		researcherDB:       researcherDB,
		timerEventCooldown: timerEventCooldown,
	}
}

func (s *Runner) Run() {
	go s.startTimerThread()
}

func (s *Runner) startTimerThread() {
	// TODO: turn of gracefully
	for {
		delay := s.timerEventCooldown
		<-time.After(time.Duration(delay) * time.Second)
		go s.CleanUpExpiredParticipantContacts()
	}
}

func (s Runner) CleanUpExpiredParticipantContacts() {
	studyInfos, err := s.researcherDB.FindAllStudyInfos()
	if err != nil {
		logger.Error.Println(err)
		return
	}
	for _, info := range studyInfos {
		logger.Info.Printf("running cleanup of expired participant contacts for %s", info.Key)
		err = s.researcherDB.CleanUpExpiredParticipantContacts(info.Key, deleteAfterInDays)
		if err != nil {
			logger.Error.Println(err)
		}
	}
}
