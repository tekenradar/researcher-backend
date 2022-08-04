package v1

import (
	"github.com/tekenradar/researcher-backend/pkg/db"
	"github.com/tekenradar/researcher-backend/pkg/grpc/clients"
	"github.com/tekenradar/researcher-backend/pkg/types"
)

type HttpEndpoints struct {
	clients                 *clients.APIClients
	researcherDB            *db.ResearcherDBService
	samlConfig              *types.SAMLConfig
	useDummyLogin           bool
	loginSuccessRedirectURL string
	apiKeys                 []string
}

func NewHTTPHandler(
	clients *clients.APIClients,
	researcherDB *db.ResearcherDBService,
	samlConfig *types.SAMLConfig,
	useDummyLogin bool,
	loginSuccessRedirectURL string,
	apiKeys []string,
) *HttpEndpoints {
	return &HttpEndpoints{
		clients:                 clients,
		researcherDB:            researcherDB,
		samlConfig:              samlConfig,
		useDummyLogin:           useDummyLogin,
		loginSuccessRedirectURL: loginSuccessRedirectURL,
		apiKeys:                 apiKeys,
	}
}
