package main

import (
	"github.com/eliona-smart-building-assistant/app-integration-tests/app"
	"github.com/eliona-smart-building-assistant/app-integration-tests/test"
	"testing"
)

func TestApp(t *testing.T) {
	app.StartApp()
	test.AppWorks(t)
	app.StopApp()
}
