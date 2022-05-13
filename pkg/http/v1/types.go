package v1

type SAMLConfig struct {
	IDPUrl                   string `yaml:"idp_root_url"`
	SPRootUrl                string `yaml:"sp_root_url"`
	EntityID                 string `yaml:"entity_id"`
	MetaDataURL              string `yaml:"metadata_url"`
	SessionCertPath          string `yaml:"session_cert"`
	SessionKeyPath           string `yaml:"session_key"`
	TemplatePathLoginSuccess string `yaml:"templates_login_success"`
}
