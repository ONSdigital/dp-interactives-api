package config

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("Given an environment with no environment variables set", t, func() {
		cfg, err := Get()

		Convey("When the config values are retrieved", func() {

			Convey("Then there should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the values should be set to the expected defaults", func() {
				So(cfg.BindAddr, ShouldEqual, "localhost:24700")
				So(cfg.ApiURL, ShouldResemble, "http://localhost:24700")
				So(cfg.Brokers, ShouldResemble, []string{"localhost:9092", "localhost:9093", "localhost:9094"})
				So(cfg.KafkaVersion, ShouldEqual, "1.0.2")
				So(cfg.KafkaSecProtocol, ShouldEqual, "")
				So(cfg.KafkaMaxBytes, ShouldEqual, 2000000)
				So(cfg.InteractivesTopic, ShouldEqual, "interactives-visualisations")
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.MongoConfig.BindAddr, ShouldEqual, "localhost:27017")
				So(cfg.MongoConfig.Collection, ShouldEqual, "visualisations")
				So(cfg.MongoConfig.Database, ShouldEqual, "interactives")
				So(cfg.MongoConfig.Username, ShouldEqual, "")
				So(cfg.MongoConfig.Password, ShouldEqual, "")
				So(cfg.MongoConfig.IsSSL, ShouldEqual, false)
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, cfg)
			})
		})
	})
}