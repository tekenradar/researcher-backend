package v1

import (
	"github.com/tekenradar/researcher-backend/pkg/db"
	"github.com/tekenradar/researcher-backend/pkg/types"
)

type HttpEndpoints struct {
	researcherDB            *db.ResearcherDBService
	samlConfig              *types.SAMLConfig
	useDummyLogin           bool
	loginSuccessRedirectURL string
	apiKeys                 []string
}

func NewHTTPHandler(
	researcherDB *db.ResearcherDBService,
	samlConfig *types.SAMLConfig,
	useDummyLogin bool,
	loginSuccessRedirectURL string,
	apiKeys []string,
) *HttpEndpoints {
	return &HttpEndpoints{
		researcherDB:            researcherDB,
		samlConfig:              samlConfig,
		useDummyLogin:           useDummyLogin,
		loginSuccessRedirectURL: loginSuccessRedirectURL,
		apiKeys:                 apiKeys,
	}
}
