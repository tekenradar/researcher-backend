package v1

type HttpEndpoints struct {
	samlConfig *SAMLConfig
}

func NewHTTPHandler(
	samlConfig *SAMLConfig,
) *HttpEndpoints {
	return &HttpEndpoints{
		samlConfig: samlConfig,
	}
}
