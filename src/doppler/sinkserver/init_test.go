package sinkserver_test

import (
	"doppler/envelopewrapper"
	"doppler/iprange"
	"doppler/sinkserver"
	"doppler/sinkserver/blacklist"
	"doppler/sinkserver/sinkmanager"
	"doppler/sinkserver/websocketserver"
	"github.com/cloudfoundry/loggregatorlib/appservice"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent"
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	"time"
)

var sinkManager *sinkmanager.SinkManager

var TestMessageRouter *sinkserver.MessageRouter
var TestWebsocketServer *websocketserver.WebsocketServer
var dataReadChannel chan *envelopewrapper.WrappedEnvelope

var blackListTestWebsocketServer *websocketserver.WebsocketServer
var blackListDataReadChannel chan *envelopewrapper.WrappedEnvelope

const (
	SERVER_PORT              = "9081"
	BLACKLIST_SERVER_PORT    = "9082"
	FAST_TIMEOUT_SERVER_PORT = "9083"
)

const SECRET = "secret"

func init() {
	dataReadChannel = make(chan *envelopewrapper.WrappedEnvelope)

	logger := loggertesthelper.Logger()
	cfcomponent.Logger = logger

	newAppServiceChan := make(chan appservice.AppService)
	deletedAppServiceChan := make(chan appservice.AppService)

	emptyBlacklist := blacklist.New(nil)
	sinkManager, _ = sinkmanager.NewSinkManager(1024, false, emptyBlacklist, logger)
	go sinkManager.Start(newAppServiceChan, deletedAppServiceChan)

	TestMessageRouter = sinkserver.NewMessageRouter(sinkManager, logger)
	go TestMessageRouter.Start(dataReadChannel)

	apiEndpoint := "localhost:" + SERVER_PORT
	TestWebsocketServer = websocketserver.New(apiEndpoint, sinkManager, 10*time.Second, 100, loggertesthelper.Logger())
	go TestWebsocketServer.Start()

	timeoutApiEndpoint := "localhost:" + FAST_TIMEOUT_SERVER_PORT
	FastTimeoutTestWebsocketServer := websocketserver.New(timeoutApiEndpoint, sinkManager, 10*time.Millisecond, 100, loggertesthelper.Logger())
	go FastTimeoutTestWebsocketServer.Start()

	blackListDataReadChannel = make(chan *envelopewrapper.WrappedEnvelope)
	localhostBlacklist := blacklist.New([]iprange.IPRange{iprange.IPRange{Start: "127.0.0.0", End: "127.0.0.2"}})
	blacklistSinkManager, _ := sinkmanager.NewSinkManager(1024, false, localhostBlacklist, logger)
	go blacklistSinkManager.Start(newAppServiceChan, deletedAppServiceChan)

	blacklistTestMessageRouter := sinkserver.NewMessageRouter(blacklistSinkManager, logger)
	go blacklistTestMessageRouter.Start(blackListDataReadChannel)

	blacklistApiEndpoint := "localhost:" + BLACKLIST_SERVER_PORT
	blackListTestWebsocketServer = websocketserver.New(blacklistApiEndpoint, blacklistSinkManager, 10*time.Second, 100, loggertesthelper.Logger())
	go blackListTestWebsocketServer.Start()

	time.Sleep(5 * time.Millisecond)
}

func WaitForWebsocketRegistration() {
	time.Sleep(50 * time.Millisecond)
}