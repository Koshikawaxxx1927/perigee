package vcs

import (
	"math"
	"time"
)

func VCSScheduler() {
	relativeError := VivaldiModels[MyEnode].Coordinate().E
	sleepTime := math.Min(10.0, inverseLog(relativeError))
	// sleepTime := 10.0
	log.Infof("Wait for %d seconds until next VCS running | RelativeError: %.2f", int(sleepTime)*60, relativeError)
	// log.Infof("Wait for %d seconds until next VCS running", int(sleepTime)*60)
	time.Sleep(time.Duration(sleepTime) * time.Minute)

	RunVCSChan <- "Run VCS"
}
