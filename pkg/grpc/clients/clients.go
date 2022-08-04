package clients

import (
	"github.com/coneno/logger"

	studyAPI "github.com/influenzanet/study-service/pkg/api"
	"google.golang.org/grpc"
)

// APIClients holds the service clients to the internal services
type APIClients struct {
	StudyService studyAPI.StudyServiceApiClient
}

func connectToGRPCServer(addr string, maxMsgSize int) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(maxMsgSize),
		grpc.MaxCallSendMsgSize(maxMsgSize),
	))
	if err != nil {
		logger.Error.Fatalf("failed to connect to %s: %v", addr, err)
	}
	return conn
}

func ConnectToStudyService(addr string, maxMsgSize int) (client studyAPI.StudyServiceApiClient, close func() error) {
	serverConn := connectToGRPCServer(addr, maxMsgSize)
	return studyAPI.NewStudyServiceApiClient(serverConn), serverConn.Close
}
