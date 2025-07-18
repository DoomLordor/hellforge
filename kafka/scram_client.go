package kafka

import (
	"github.com/xdg-go/scram"
)

// XDGSCRAMClient representation
type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

// Begin constructs a client-side authentication conversation
func (x *XDGSCRAMClient) Begin(userName, password, authzID string) error {
	var err error
	x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.ClientConversation = x.Client.NewConversation()

	return nil
}

// Step returns a string to be sent to the server or an error if the server message is invalid
func (x *XDGSCRAMClient) Step(challenge string) (string, error) {
	return x.ClientConversation.Step(challenge)
}

// Done returns true if the conversation is completed or has errored
func (x *XDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
}
