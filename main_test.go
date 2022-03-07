package main

import (
	"context"
	"flag"
	"os"
	"testing"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/features/steps"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

type componenttestSuite struct {
	MongoFeature *componenttest.MongoFeature
}

var componentFlag = flag.Bool("component", false, "perform component tests")

func (c *componenttestSuite) InitializeScenario(ctx *godog.ScenarioContext) {
	interactivesFtr, err := steps.NewInteractivesApiComponent(c.MongoFeature.Server.URI())
	if err != nil {
		panic(err)
	}

	apiFeature := componenttest.NewAPIFeature(interactivesFtr.InitialiseService)
	interactivesFtr.ApiFeature = apiFeature

	ctx.BeforeScenario(func(*godog.Scenario) {
		apiFeature.Reset()
		if err := interactivesFtr.Reset(); err != nil {
			panic(err)
		}
		if err := c.MongoFeature.Reset(); err != nil {
			log.Error(context.Background(), "failed to reset mongo feature", err)
		}
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {
		interactivesFtr.Close()
	})

	interactivesFtr.RegisterSteps(ctx)
	apiFeature.RegisterSteps(ctx)
}

func (t *componenttestSuite) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	cfg, _ := config.Get()
	ctx.BeforeSuite(func() {
		mongoOptions := componenttest.MongoOptions{
			MongoVersion: "4.4.8",
			DatabaseName: cfg.MongoConfig.Database,
		}
		t.MongoFeature = componenttest.NewMongoFeature(mongoOptions)
	})

	ctx.AfterSuite(func() {
		t.MongoFeature.Close()
	})
}

func TestComponent(t *testing.T) {
	if *componentFlag {
		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Paths:  flag.Args(),
			Format: "pretty",
		}

		ts := &componenttestSuite{}

		status := godog.TestSuite{
			Name:                 "component_tests",
			ScenarioInitializer:  ts.InitializeScenario,
			TestSuiteInitializer: ts.InitializeTestSuite,
			Options:              &opts,
		}.Run()

		if status > 0 {
			t.Fail()
		}
	} else {
		t.Skip()
	}
}
