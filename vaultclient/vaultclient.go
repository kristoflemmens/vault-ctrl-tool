package vaultclient

import (
	"github.com/hashicorp/vault/api"
	"github.com/hootsuite/vault-ctrl-tool/v2/config"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

const SecretsServicePathV1 = "/secret/application-config/services/"
const SecretsServicePathV2 = "/kv/data/application-config/services/"

type VaultClient interface {
	VerifyVaultToken(vaultToken string) (*api.Secret, error)
	Delegate() *api.Client
	FetchAWSSTSCredential(awsConfig config.AWSType) (*AWSSTSCredential, *api.Secret, error)
	CreateSSHCertificate(sshConfig config.SSHCertificateType) error
	RefreshVaultToken() (*api.Secret, error)
	ServiceSecretPrefix(configVersion int) string
}

type wrappedVaultClient struct {
	delegate *api.Client

	log zerolog.Logger
}

func NewVaultClient() (VaultClient, error) {

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}

	log := zlog.With().Str("vaultAddr", client.Address()).Logger()

	return &wrappedVaultClient{
		delegate: client,
		log:      log,
	}, nil
}

func (vc *wrappedVaultClient) Delegate() *api.Client {
	return vc.delegate
}

func (vc *wrappedVaultClient) VerifyVaultToken(vaultToken string) (*api.Secret, error) {
	vc.log.Debug().Msg("verifying vault token")
	oldToken := vc.delegate.Token()
	defer vc.delegate.SetToken(oldToken)

	vc.delegate.SetToken(vaultToken)
	secret, err := vc.delegate.Auth().Token().LookupSelf()
	if err != nil {
		vc.log.Debug().Err(err).Msg("verification failed")
		return nil, err
	}
	vc.log.Debug().Msg("verification successful")
	return secret, nil
}

func (vc *wrappedVaultClient) RefreshVaultToken() (*api.Secret, error) {
	vc.log.Debug().Msg("refreshing period on vault token")
	return vc.Delegate().Auth().Token().RenewSelf(86400) // this value is basically ignored by the server
}

func (vc *wrappedVaultClient) ServiceSecretPrefix(configVersion int) string {
	if configVersion < 2 {
		return SecretsServicePathV1
	} else {
		return SecretsServicePathV2
	}
}