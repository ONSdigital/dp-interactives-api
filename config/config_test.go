package config

import (
	"reflect"
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
				So(cfg.BindAddr, ShouldEqual, ":27500")
				So(cfg.Brokers, ShouldResemble, []string{"localhost:9092"})
				So(cfg.KafkaVersion, ShouldEqual, "1.0.2")
				So(cfg.KafkaSecProtocol, ShouldEqual, "")
				So(cfg.KafkaMaxBytes, ShouldEqual, 2000000)
				So(cfg.InteractivesWriteTopic, ShouldEqual, "interactives-import")
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.DefaultLimit, ShouldEqual, 20)
				So(cfg.DefaultMaxLimit, ShouldEqual, 100)
				So(cfg.DefaultOffset, ShouldEqual, 0)
				So(cfg.MongoConfig.ClusterEndpoint, ShouldEqual, "localhost:27017")
				So(cfg.MongoConfig.Database, ShouldEqual, "interactives")
				So(cfg.MongoConfig.Username, ShouldEqual, "")
				So(cfg.MongoConfig.Password, ShouldEqual, "")
				So(cfg.MongoConfig.IsSSL, ShouldEqual, false)
				So(cfg.AuthorisationConfig, ShouldNotBeNil)
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, cfg)
			})

			Convey("Sensitive fields are hidden", func() {
				tagName := "json"
				tagValue := "-"
				newCfg, newErr := Get()

				So(newErr, ShouldBeNil)

				sensitiveFields := getStructFieldName(newCfg, &newCfg.ServiceAuthToken, &newCfg.KafkaSecClientKey)
				for _, fld := range sensitiveFields {
					field, ok := reflect.TypeOf(cfg).Elem().FieldByName(fld)
					So(ok, ShouldBeTrue)
					So(field.Tag.Get(tagName), ShouldEqual, tagValue)
				}
			})
		})
	})
}

func getStructFieldName(Struct interface{}, StructField ...interface{}) (fields []string) {
	fields = []string{}

	for r := range StructField {
		s := reflect.ValueOf(Struct).Elem()
		f := reflect.ValueOf(StructField[r]).Elem()

		for i := 0; i < s.NumField(); i++ {
			valueField := s.Field(i)
			if valueField.Addr().Interface() == f.Addr().Interface() {
				fields = append(fields, s.Type().Field(i).Name)
			}
		}
	}
	return fields
}
