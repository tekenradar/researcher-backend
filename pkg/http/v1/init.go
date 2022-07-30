package v1

import "github.com/tekenradar/researcher-backend/pkg/types"

type HttpEndpoints struct {
	samlConfig              *types.SAMLConfig
	useDummyLogin           bool
	loginSuccessRedirectURL string
	apiKeys                 []string
}

func NewHTTPHandler(
	samlConfig *types.SAMLConfig,
	useDummyLogin bool,
	loginSuccessRedirectURL string,
	apiKeys []string,
) *HttpEndpoints {
	return &HttpEndpoints{
		samlConfig:              samlConfig,
		useDummyLogin:           useDummyLogin,
		loginSuccessRedirectURL: loginSuccessRedirectURL,
		apiKeys:                 apiKeys,
	}
}
