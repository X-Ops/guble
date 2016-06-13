package server

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/smancke/guble/protocol"
	"github.com/smancke/guble/server/webserver"

	"github.com/docker/distribution/health"
	"github.com/smancke/guble/metrics"
	"net/http"
	"reflect"
	"time"
)

const (
	defaultHealthFrequency = time.Second * 60
	defaultHealthThreshold = 1
)

var loggerService = log.WithFields(log.Fields{
	"app":    "guble",
	"module": "service",
	"env":    "TBD"})

// Startable interface for modules which provide a start mechanism
type Startable interface {
	Start() error
}

// Stopable interface for modules which provide a stop mechanism
type Stopable interface {
	Stop() error
}

// Endpoint adds a HTTP handler for the `GetPrefix()` to the webserver
type Endpoint interface {
	http.Handler
	GetPrefix() string
}

// Service is the main class for simple control of a server
type Service struct {
	webserver       *webserver.WebServer
	router          Router
	modules         []interface{}
	healthEndpoint  string
	healthFrequency time.Duration
	healthThreshold int
	metricsEndpoint string
	gubleNodeID     string
	gubleNodesURLs  []string
}

// NewService registers the Main Router, where other modules can subscribe for messages
func NewService(router Router, webserver *webserver.WebServer) *Service {
	service := &Service{
		webserver:       webserver,
		router:          router,
		healthFrequency: defaultHealthFrequency,
		healthThreshold: defaultHealthThreshold,
	}
	service.RegisterModules(service.router, service.webserver)
	return service
}

func (s *Service) RegisterModules(modules ...interface{}) {
	loggerService.WithFields(log.Fields{
		"numberOfNewModules":       len(s.modules),
		"numberOfExsistingModules": len(modules),
	}).Debug(" RegisterModules: adding")

	s.modules = append(s.modules, modules...)
}

func (s *Service) HealthEndpointPrefix(value string) *Service {
	s.healthEndpoint = value
	return s
}

func (s *Service) MetricsEndpointPrefix(value string) *Service {
	s.metricsEndpoint = value
	return s
}

func (s *Service) GubleNodeID(value string) *Service {
	logger.WithField("NodeID", value).Info("Setting node-id")
	s.gubleNodeID = value
	return s
}

func (s *Service) GubleNodesURLs(values []string) *Service {
	if s.gubleNodeID != "" {
		logger.WithField("peers", values).Info("Setting peer nodes")
		s.gubleNodesURLs = values
	}
	return s
}

// Start checks the modules for the following interfaces and registers and/or starts:
//   Startable:
//   health.Checker:
//   Endpoint: Register the handler function of the Endpoint in the http service at prefix
func (s *Service) Start() error {
	el := protocol.NewErrorList("service: errors occured while starting: ")

	if s.healthEndpoint != "" {
		logger.WithField("healthEndpoint", s.healthEndpoint).Info("Health endpoint")
		s.webserver.Handle(s.healthEndpoint, http.HandlerFunc(health.StatusHandler))
	} else {
		logger.Debug("Health endpoint disabled")
	}

	if s.metricsEndpoint != "" {
		logger.WithField("metricsEndpoint", s.metricsEndpoint).Info("Metrics Endpoint")
		s.webserver.Handle(s.metricsEndpoint, http.HandlerFunc(metrics.HttpHandler))
	} else {
		logger.Debug("Metrics endpoint disabled")
	}

	for _, module := range s.modules {
		name := reflect.TypeOf(module).String()
		if startable, ok := module.(Startable); ok {
			loggerService.WithFields(log.Fields{
				"name": name,
			}).Info("Starting module")

			if err := startable.Start(); err != nil {

				loggerService.WithFields(log.Fields{
					"name": name,
					"err":  err,
				}).Error("Error while starting module")

				el.Add(err)
			}
		}
		if checker, ok := module.(health.Checker); ok && s.healthEndpoint != "" {

			logger.WithField("name", name).Info("Registering module as HealthChecker")
			health.RegisterPeriodicThresholdFunc(name, s.healthFrequency, s.healthThreshold, health.CheckFunc(checker.Check))
		}
		if endpoint, ok := module.(Endpoint); ok {
			prefix := endpoint.GetPrefix()
			loggerService.WithFields(log.Fields{
				"name":   name,
				"prefix": prefix,
			}).Info("Resgistering Endpoint  module with name")
			s.webserver.Handle(prefix, endpoint)
		}
	}
	return el.ErrorOrNil()
}

// Stop stops the registered modules in the required order
func (s *Service) Stop() error {
	var stopables []Stopable

	for _, module := range s.modules {
		if stopable, ok := module.(Stopable); ok {
			stopables = append(stopables, stopable)
		}
	}

	stopOrder := make([]int, len(stopables))
	for i := 1; i < len(stopables); i++ {
		stopOrder[i] = len(stopables) - i
	}
	loggerService.WithFields(log.Fields{
		"numberOfNewModules":       len(stopOrder),
		"numberOfExsistingModules": stopOrder,
	}).Debug("Stopping modules in this order relative to registration")

	errors := protocol.NewErrorList("stopping errors: ")
	for _, order := range stopOrder {
		module := stopables[order]
		name := reflect.TypeOf(module).String()

		loggerService.WithFields(log.Fields{
			"name":  name,
			"order": order,
		}).Info("Stopping module with name")
		if err := module.Stop(); err != nil {
			errors.Add(err)
		}
	}

	if err := errors.ErrorOrNil(); err != nil {
		return fmt.Errorf("service: errors while stopping modules: %s", err)
	}

	return nil
}

// Modules returns the slice of registered modules
func (s *Service) Modules() []interface{} {
	return s.modules
}

// WebServer returns the service *webserver.WebServer instance
func (s *Service) WebServer() *webserver.WebServer {
	return s.webserver
}
