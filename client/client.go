package client

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	gonetconf "github.com/openshift-telco/go-netconf/netconf"
	gonetconfmsg "github.com/openshift-telco/go-netconf/netconf/message"
	"golang.org/x/crypto/ssh"
	"time"
)

type NetconfClientFactory interface {
	Create(logger logr.Logger, targetHost string, targetPort int, username string, password string) NetconfClient
}

type NetconfClientFactoryBroker struct {
}

func (r NetconfClientFactoryBroker) Create(logger logr.Logger, targetHost string, targetPort int, username string, password string) NetconfClient {
	return &NetconfClientBroker{logger: logger, targetHost: targetHost, targetPort: targetPort, username: username, password: password}
}

var DefaultNetconfClientFactoryBroker NetconfClientFactory = &NetconfClientFactoryBroker{}

type NetconfClient interface {
	ExecEditConfig(config interface{}) error
}

type NetconfClientBroker struct {
	logger     logr.Logger
	targetHost string
	targetPort int
	username   string
	password   string
}

func (r *NetconfClientBroker) ExecEditConfig(config interface{}) error {
	configData, err := xml.Marshal(&config)
	if err != nil {
		return err
	}
	r.logger.Info("Netconf edit-config", "configData", string(configData))
	e := gonetconfmsg.NewEditConfig(gonetconfmsg.DatastoreCandidate, gonetconfmsg.DefaultOperationTypeMerge, string(configData))

	session, err := r.createSession()
	if err != nil {
		return err
	}
	defer func() {
		_ = session.Close()
	}()

	if err := session.AsyncRPC(e, r.defaultLogRpcReplyCallback(e.MessageID)); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (r *NetconfClientBroker) createSession() (*gonetconf.Session, error) {
	sshConfig := &ssh.ClientConfig{
		User:            r.username,
		Auth:            []ssh.AuthMethod{ssh.Password(r.password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	s, err := gonetconf.DialSSH(fmt.Sprintf("%s:%d", r.targetHost, r.targetPort), sshConfig)
	if err != nil {
		return nil, err
	}
	capabilities := gonetconf.DefaultCapabilities
	err = s.SendHello(&gonetconfmsg.Hello{Capabilities: capabilities})
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (r *NetconfClientBroker) defaultLogRpcReplyCallback(eventId string) gonetconf.Callback {
	return func(event gonetconf.Event) {
		reply := event.RPCReply()
		if reply == nil {
			r.logger.Error(errors.New("Netconf nil reply"), "Failed to execute RPC")
		}
		if event.EventID() == eventId {
			r.logger.Info("Successfully executed RPC", "rawReply", reply.RawReply)
			println(reply.RawReply)
		}
	}
}
