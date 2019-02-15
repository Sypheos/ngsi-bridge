package bridges

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"ngsi-bridge/ngsi"

	ttnSdk "github.com/TheThingsNetwork/go-app-sdk"
	"github.com/TheThingsNetwork/go-utils/log"
	ttnTypes "github.com/TheThingsNetwork/ttn/core/types"
	"github.com/pkg/errors"
)

// TTNBridge for Waternet.
type TTNBridge struct {
	ctx        log.Interface
	Ttn        TtnAccess
	ClientName string
	client     ttnSdk.Client
	pubSub     ttnSdk.ApplicationPubSub
	work       func(up *ttnTypes.UplinkMessage)
	schemas    map[string]*Schema
	broker     string
}

// TtnAccess value to access TTN
type TtnAccess struct {
	AppID           string `mapstructure:"app-id"`
	AppKey          string `mapstructure:"app-key"`
	AccountServer   string `mapstructure:"account-server"`
	DiscoveryServer string `mapstructure:"discovery-server"`
	CaCert          string `mapstructure:"ca-cert"`
}

// StartMQTT create and start a TTNBridge on the community network. Subribe to all the device uplink message and start
// a go routine that listen for incoming uplink to send them to Fiware.
func (m *TTNBridge) Prepare(ctx log.Interface, mapperSchema map[string]*Schema, broker string) error {
	config := ttnSdk.NewConfig(m.ClientName, m.Ttn.AccountServer, m.Ttn.DiscoveryServer)
	m.ctx = ctx.WithField("endpoint", "MQTT")
	m.ctx.Info("Building bridge...")
	if caCert := m.Ttn.CaCert; caCert != "" {
		config.TLSConfig = new(tls.Config)
		certBytes, err := ioutil.ReadFile(caCert)
		if err != nil {
			return err
		}
		config.TLSConfig.RootCAs = x509.NewCertPool()
		if ok := config.TLSConfig.RootCAs.AppendCertsFromPEM(certBytes); !ok {
			return errors.New("could not use CA certificate")
		}
	}
	m.schemas = mapperSchema
	m.client = config.NewClient(m.Ttn.AppID, m.Ttn.AppKey)
	m.work = m.handleUp
	m.ctx.Info("Bridge built.")
	return nil
}

// Close close the resources and MQTT connection
func (m *TTNBridge) Close() error {
	m.ctx.Info("Closing bridge.")
	m.pubSub.Close()
	return m.client.Close()
}

func (m *TTNBridge) Open() error {
	m.ctx.Info("Opening bridge...")
	var err error
	m.pubSub, err = m.client.PubSub()
	if err != nil {
		return err
	}
	m.ctx.Info("Pubsub")
	devices := m.pubSub.AllDevices()
	up, err := devices.SubscribeUplink()
	if err != nil {
		return err
	}
	m.ctx.Info("Bridging complete.")
	for uplink := range up {
		m.work(uplink)
	}
	m.ctx.Info("Bridging closed.")
	return nil
}

func (m *TTNBridge) handleUp(up *ttnTypes.UplinkMessage) {
	t, ok := up.Attributes["type"]
	if !ok {
		m.ctx.Debug("No type attribute using 'ttn'.")
		t = "ttn"
	}
	sch, ok := m.schemas[t]
	if !ok {
		m.ctx.Warnf("No schema defined for type %s", t)
		return
	}
	ent, err := decode(up.PayloadFields, sch)
	if err != nil {
		m.ctx.WithError(err).Warn("Could not decode uplink.")
		return
	}
	err = ngsi.PushAttributes(m.ctx, m.broker, ent)
	if err != nil {
		m.ctx.WithError(err).Warn("Could not push entity to broker.")
		return
	}
}
