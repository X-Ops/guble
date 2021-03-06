package router

import (
	"fmt"
	"runtime"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/distribution/health"

	"encoding/json"

	"github.com/smancke/guble/protocol"
	"github.com/smancke/guble/server/auth"
	"github.com/smancke/guble/server/cluster"
	"github.com/smancke/guble/server/kvstore"
	"github.com/smancke/guble/server/store"
)

const (
	overloadedHandleChannelRatio = 0.9
	handleChannelCapacity        = 500
	subscribeChannelCapacity     = 10
	unsubscribeChannelCapacity   = 10
)

// Router interface provides a mechanism for PubSub messaging
type Router interface {
	AccessManager() (auth.AccessManager, error)
	MessageStore() (store.MessageStore, error)
	KVStore() (kvstore.KVStore, error)
	Cluster() *cluster.Cluster
	Fetch(*store.FetchRequest) error

	Subscribe(r *Route) (*Route, error)
	Unsubscribe(r *Route)
	HandleMessage(message *protocol.Message) error
	GetSubscribersForTopic(topic string) ([]byte, error)
	Done() <-chan bool
}

// Helper struct to pass `Route` to subscription channel and provide a notification channel.
type subRequest struct {
	route *Route
	doneC chan bool
}

type router struct {
	routes       map[protocol.Path][]*Route // mapping the path to the route slice
	handleC      chan *protocol.Message
	subscribeC   chan subRequest
	unsubscribeC chan subRequest
	stopC        chan bool      // Channel that signals stop of the router
	stopping     bool           // Flag: the router is in stopping process and no incoming messages are accepted
	wg           sync.WaitGroup // Add any operation that we need to wait upon here

	accessManager auth.AccessManager
	messageStore  store.MessageStore
	kvStore       kvstore.KVStore
	cluster       *cluster.Cluster

	sync.RWMutex
}

// New returns a pointer to Router
func New(accessManager auth.AccessManager, messageStore store.MessageStore, kvStore kvstore.KVStore, cluster *cluster.Cluster) Router {
	return &router{
		routes: make(map[protocol.Path][]*Route),

		handleC:      make(chan *protocol.Message, handleChannelCapacity),
		subscribeC:   make(chan subRequest, subscribeChannelCapacity),
		unsubscribeC: make(chan subRequest, unsubscribeChannelCapacity),
		stopC:        make(chan bool, 1),

		accessManager: accessManager,
		messageStore:  messageStore,
		kvStore:       kvStore,
		cluster:       cluster,
	}
}

func (router *router) Start() error {
	router.panicIfInternalDependenciesAreNil()
	logger.Info("Starting router")
	resetRouterMetrics()

	router.wg.Add(1)
	router.setStopping(false)

	go func() {
		for {
			if router.stopping && router.channelsAreEmpty() {
				router.closeRoutes()
				router.wg.Done()
				return
			}

			func() {
				defer protocol.PanicLogger()

				select {
				case message := <-router.handleC:
					router.routeMessage(message)
					runtime.Gosched()
				case subscriber := <-router.subscribeC:
					router.subscribe(subscriber.route)
					subscriber.doneC <- true
				case unsubscriber := <-router.unsubscribeC:
					router.unsubscribe(unsubscriber.route)
					unsubscriber.doneC <- true
				case <-router.Done():
					router.setStopping(true)
				}
			}()
		}
	}()

	return nil
}

// Stop stops the router by closing the stop channel, and waiting on the WaitGroup
func (router *router) Stop() error {
	logger.Info("Stopping router")

	router.stopC <- true
	router.wg.Wait()
	return nil
}

func (router *router) Check() error {
	if router.accessManager == nil || router.messageStore == nil || router.kvStore == nil {
		logger.WithError(ErrServiceNotProvided).Error("Some mandatory services are not provided")
		return ErrServiceNotProvided
	}
	if checkable, ok := router.messageStore.(health.Checker); ok {
		err := checkable.Check()
		if err != nil {
			logger.WithError(err).Error("MessageStore check failed")
			return err
		}
	}
	if checkable, ok := router.kvStore.(health.Checker); ok {
		err := checkable.Check()
		if err != nil {
			logger.WithError(err).Error("KVStore check failed")
			return err
		}
	}
	return nil
}

// HandleMessage stores the message in the MessageStore(and gets a new ID for it if the message was created locally)
// and then passes it to the internal channel, and asynchronously to the cluster (if available).
func (router *router) HandleMessage(message *protocol.Message) error {
	logger.WithFields(log.Fields{
		"userID": message.UserID,
		"path":   message.Path}).Debug("HandleMessage")

	mTotalMessagesIncoming.Add(1)
	if err := router.isStopping(); err != nil {
		logger.WithError(err).Error("Router is stopping")
		return err
	}

	if !router.accessManager.IsAllowed(auth.WRITE, message.UserID, message.Path) {
		return &PermissionDeniedError{UserID: message.UserID, AccessType: auth.WRITE, Path: message.Path}
	}

	var nodeID uint8
	if router.cluster != nil {
		nodeID = router.cluster.Config.ID
	}

	mTotalMessagesIncomingBytes.Add(int64(len(message.Bytes())))
	size, err := router.messageStore.StoreMessage(message, nodeID)
	if err != nil {
		logger.WithError(err).Error("Error storing message")
		mTotalMessageStoreErrors.Add(1)
		return err
	}
	mTotalMessagesStoredBytes.Add(int64(size))

	router.handleOverloadedChannel()

	router.handleC <- message

	if router.cluster != nil && message.NodeID == router.cluster.Config.ID {
		go router.cluster.BroadcastMessage(message)
	}

	return nil
}

func (router *router) GetSubscribersForTopic(topicPath string) ([]byte, error) {
	subscribers := make([]RouteParams, 0)
	routes, present := router.routes[protocol.Path(topicPath)]
	if present {
		for index, currRoute := range routes {
			logger.WithFields(log.Fields{
				"index":       index,
				"routeParams": currRoute.RouteParams,
			}).Debug("Added route to slice")
			subscribers = append(subscribers, currRoute.RouteParams)
		}
	}
	return json.Marshal(subscribers)
}

// Subscribe adds a route to the subscribers. If there is already a route with same Application Id and Path, it will be replaced.
func (router *router) Subscribe(r *Route) (*Route, error) {
	logger.WithFields(log.Fields{
		"accessManager": router.accessManager,
		"route":         r,
	}).Debug("Subscribe")

	if err := router.isStopping(); err != nil {
		return nil, err
	}

	userID := r.Get("user_id")
	routePath := r.Path

	accessAllowed := router.accessManager.IsAllowed(auth.READ, userID, routePath)
	if !accessAllowed {
		return r, &PermissionDeniedError{UserID: userID, AccessType: auth.READ, Path: routePath}
	}
	req := subRequest{
		route: r,
		doneC: make(chan bool),
	}

	router.subscribeC <- req
	<-req.doneC
	return r, nil
}

func (router *router) Unsubscribe(r *Route) {
	logger.WithFields(log.Fields{
		"accessManager": router.accessManager,
		"route":         r,
	}).Debug("Unsubscribe")

	req := subRequest{
		route: r,
		doneC: make(chan bool),
	}
	router.unsubscribeC <- req
	<-req.doneC
}

func (router *router) subscribe(r *Route) {
	logger.WithField("route", r).Debug("Internal subscribe")
	mTotalSubscriptionAttempts.Add(1)

	routePath := r.Path
	slice, present := router.routes[routePath]
	var removed bool
	if present {
		// Try to remove, to avoid double subscriptions of the same app
		slice, removed = removeIfMatching(slice, r)
	} else {
		// Path not present yet. Initialize the slice
		slice = make([]*Route, 0, 1)
		router.routes[routePath] = slice
		mCurrentRoutes.Add(1)
	}
	router.routes[routePath] = append(slice, r)
	if removed {
		mTotalDuplicateSubscriptionsAttempts.Add(1)
	} else {
		mTotalSubscriptions.Add(1)
		mCurrentSubscriptions.Add(1)
	}
}

func (router *router) unsubscribe(r *Route) {
	logger.WithField("route", r).Debug("Internal unsubscribe")
	mTotalUnsubscriptionAttempts.Add(1)

	routePath := r.Path
	slice, present := router.routes[routePath]
	if !present {
		mTotalInvalidTopicOnUnsubscriptionAttempts.Add(1)
		return
	}
	var removed bool
	router.routes[routePath], removed = removeIfMatching(slice, r)
	if removed {
		mTotalUnsubscriptions.Add(1)
		mCurrentSubscriptions.Add(-1)
	} else {
		mTotalInvalidUnsubscriptionAttempts.Add(1)
	}
	if len(router.routes[routePath]) == 0 {
		delete(router.routes, routePath)
		mCurrentRoutes.Add(-1)
	}
}

func (router *router) panicIfInternalDependenciesAreNil() {
	if router.accessManager == nil || router.kvStore == nil || router.messageStore == nil {
		panic(fmt.Sprintf("router: the internal dependencies marked with `true` are not set: AccessManager=%v, KVStore=%v, MessageStore=%v",
			router.accessManager == nil, router.kvStore == nil, router.messageStore == nil))
	}
}

func (router *router) channelsAreEmpty() bool {
	return len(router.handleC) == 0 && len(router.subscribeC) == 0 && len(router.unsubscribeC) == 0
}

func (router *router) setStopping(v bool) {
	router.Lock()
	defer router.Unlock()

	router.stopping = v
}

func (router *router) Done() <-chan bool {
	return router.stopC
}

func (router *router) isStopping() error {
	router.RLock()
	defer router.RUnlock()

	if router.stopping {
		return &ModuleStoppingError{"Router"}
	}

	return nil
}

func (router *router) routeMessage(message *protocol.Message) {
	logger.WithField("msgMetadata", message.Metadata()).Debug("Called routeMessage for data")
	mTotalMessagesRouted.Add(1)

	matched := false
	for path, pathRoutes := range router.routes {
		if matchesTopic(message.Path, path) {
			matched = true
			for _, route := range pathRoutes {
				if err := route.Deliver(message); err == ErrInvalidRoute {
					// Unsubscribe invalid routes
					router.unsubscribe(route)
				}
			}
		}
	}

	if !matched {
		logger.WithField("topic", message.Path).Debug("No route matched.")
		mTotalMessagesNotMatchingTopic.Add(1)
	}
}

func (router *router) closeRoutes() {
	logger.Debug("Called closeRoutes")

	for _, currentRouteList := range router.routes {
		for _, route := range currentRouteList {
			router.unsubscribe(route)
			log.WithFields(log.Fields{"module": "router", "route": route.String()}).Debug("Closing route")
			route.Close()
		}
	}
}

func (router *router) handleOverloadedChannel() {
	if float32(len(router.handleC))/float32(cap(router.handleC)) > overloadedHandleChannelRatio {
		logger.WithFields(log.Fields{
			"currentLength": len(router.handleC),
			"maxCapacity":   cap(router.handleC),
		}).Warn("handleC channel is almost full")
		mTotalOverloadedHandleChannel.Add(1)
	}
}

// matchesTopic checks whether the supplied routePath matches the message topic
func matchesTopic(messagePath, routePath protocol.Path) bool {
	messagePathLen := len(string(messagePath))
	routePathLen := len(string(routePath))
	return strings.HasPrefix(string(messagePath), string(routePath)) &&
		(messagePathLen == routePathLen ||
			(messagePathLen > routePathLen && string(messagePath)[routePathLen] == '/'))
}

// removeIfMatching removes a route from the supplied list, based on same ApplicationID id and same path (if existing)
// returns: the (possibly updated) slide, and a boolean value (true if route was removed, false otherwise)
func removeIfMatching(slice []*Route, route *Route) ([]*Route, bool) {
	position := -1
	for p, r := range slice {
		if r.Equal(route) {
			position = p
			break
		}
	}
	if position == -1 {
		return slice, false
	}
	return append(slice[:position], slice[position+1:]...), true
}

// AccessManager returns the `accessManager` provided for the router
func (router *router) AccessManager() (auth.AccessManager, error) {
	if router.accessManager == nil {
		return nil, ErrServiceNotProvided
	}
	return router.accessManager, nil
}

// MessageStore returns the `messageStore` provided for the router
func (router *router) MessageStore() (store.MessageStore, error) {
	if router.messageStore == nil {
		return nil, ErrServiceNotProvided
	}
	return router.messageStore, nil
}

// KVStore returns the `kvStore` provided for the router
func (router *router) KVStore() (kvstore.KVStore, error) {
	if router.kvStore == nil {
		return nil, ErrServiceNotProvided
	}
	return router.kvStore, nil
}

func (router *router) Fetch(req *store.FetchRequest) error {
	if err := router.isStopping(); err != nil {
		return err
	}
	router.messageStore.Fetch(req)
	return nil
}

// Cluster returns the `cluster` provided for the router, or nil if no cluster was set-up
func (router *router) Cluster() *cluster.Cluster {
	return router.cluster
}
