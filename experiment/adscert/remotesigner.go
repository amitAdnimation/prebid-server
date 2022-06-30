package adscert

import (
	"errors"
	"fmt"
	"github.com/IABTechLab/adscert/pkg/adscert/api"
	"github.com/IABTechLab/adscert/pkg/adscert/signatory"
	"github.com/prebid/prebid-server/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/url"
	"time"
)

// remoteSigner - holds the signatory to add adsCert header to requests using remote signing server
type remoteSigner struct {
	signatory signatory.AuthenticatedConnectionsSignatory
}

// Sign - adds adsCert header to requests using remote signing server
func (rs *remoteSigner) Sign(destinationURL string, body []byte) (string, error) {
	// Request the signature.
	signatureResponse, err := rs.signatory.SignAuthenticatedConnection(
		&api.AuthenticatedConnectionSignatureRequest{
			RequestInfo: createRequestInfo(destinationURL, []byte(body)),
		})
	if err != nil {
		return "", err
	}
	return getSignatureMessage(signatureResponse)
}

func newRemoteSigner(remoteSignerConfig config.AdsCertRemote) (*remoteSigner, error) {
	if err := validateRemoteSignerConfig(remoteSignerConfig); err != nil {
		return nil, err
	}
	// Establish the gRPC connection that the client will use to connect to the
	// signatory server.  This basic example uses unauthenticated connections
	// which should not be used in a production environment.
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(remoteSignerConfig.Url, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial remote signer: %v", err)
	}

	clientOpts := &signatory.AuthenticatedConnectionsSignatoryClientOptions{
		Timeout: time.Duration(remoteSignerConfig.SigningTimeoutMs) * time.Millisecond}
	signatoryClient := signatory.NewAuthenticatedConnectionsSignatoryClient(conn, clientOpts)
	return &remoteSigner{signatory: signatoryClient}, nil

}

func validateRemoteSignerConfig(remoteSignerConfig config.AdsCertRemote) error {
	_, err := url.ParseRequestURI(remoteSignerConfig.Url)
	if err != nil {
		return errors.New("invalid url for remote signer")
	}
	if remoteSignerConfig.SigningTimeoutMs <= 0 {
		return errors.New("invalid signing timeout for remote signer")
	}
	return nil
}