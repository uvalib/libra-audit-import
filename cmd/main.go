package main

import (
	"flag"
	"fmt"
	"github.com/uvalib/librabus-sdk/uvalibrabus"
	"os"
	"strings"
)

// global logging level
var logLevel string

// main entry point
func main() {

	var eventBus string
	var eventSource string
	var namespace string
	var inFile string
	var importMode string
	var dryRun bool
	var limit int

	flag.StringVar(&eventBus, "bus", "", "Event bus name")
	flag.StringVar(&eventSource, "source", "", "Event source name")
	flag.StringVar(&namespace, "namespace", "", "Namespace to import")
	flag.StringVar(&importMode, "importmode", "", "Import mode, either etd or open")
	flag.StringVar(&inFile, "infile", "", "Import file")
	flag.BoolVar(&dryRun, "dryrun", false, "Process but do not actually import")
	flag.IntVar(&limit, "limit", 0, "Number of items to import, 0 for no limit")
	flag.StringVar(&logLevel, "loglevel", "E", "Logging level (D|I|W|E)")
	flag.Parse()

	// validate required parameters
	if len(eventBus) == 0 ||
		len(eventSource) == 0 ||
		len(namespace) == 0 ||
		len(importMode) == 0 ||
		len(inFile) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if fileExists(inFile) == false {
		logError("import file does not exist or is not readable")
		os.Exit(1)
	}

	if importMode != "etd" && importMode != "open" {
		logError("import mode must be etd|open")
		os.Exit(1)
	}

	if logLevel != "D" && logLevel != "I" && logLevel != "W" && logLevel != "E" {
		logError("logging level must be D|I|W|E")
		os.Exit(1)
	}

	var bus uvalibrabus.UvaBus
	var err error

	if dryRun == false {
		// create the bus client
		cfg := uvalibrabus.UvaBusConfig{
			Source:  eventSource,
			BusName: eventBus,
			//Log:     logger,
		}
		bus, err = uvalibrabus.NewUvaBus(cfg)
		if err != nil {
			logError(fmt.Sprintf("creating event bus client (%s)", err.Error()))
			os.Exit(1)
		}
		logAlways(fmt.Sprintf("using: %s@%s", eventSource, eventBus))
	} else {
		logAlways("dryrun, NO import!!")
	}

	okCount := 0
	errCount := 0
	ignoreCount := 0

	buf, err := loadFile(inFile)
	if err != nil {
		logError(err.Error())
		os.Exit(1)
	}

	// go through our list
	lines := strings.Split(string(buf), "\n")
	for _, i := range lines {

		// ignore blank lines
		if len(i) == 0 {
			continue
		}

		// if we are limiting our import count
		if limit != 0 && ((okCount + errCount + ignoreCount) >= limit) {
			logDebug(fmt.Sprintf("terminating after %d events(s)", limit))
			break
		}

		var event *uvalibrabus.UvaBusEvent
		if importMode == "etd" {
			event, err = makeEtdEvent(namespace, i)
		} else {
			event, err = makeOpenEvent(namespace, i)
		}

		if err != nil {
			logError(fmt.Sprintf("creating event (%s), continuing", err.Error()))
			errCount++
			continue
		}

		// no event means we ignore this
		if event == nil {
			logWarning(fmt.Sprintf("ignoring event (%s)", i))
			ignoreCount++
			continue
		}

		// if we are configured to import
		if dryRun == false {
			err = bus.PublishEvent(event)

			if err != nil {
				logError(fmt.Sprintf("publishing event (%s), continuing", err.Error()))
				errCount++
				continue
			}
			logInfo(fmt.Sprintf("published: %s", event.String()))
		} else {
			logDebug(fmt.Sprintf("would publish: %s", event.String()))
		}

		okCount++
	}

	verb := "imported"
	if dryRun == true {
		verb = "processed"
	}
	logAlways(fmt.Sprintf("terminate normally, %s %d object(s), %d ignored and %d error(s)", verb, okCount, ignoreCount, errCount))
}

//
// end of file
//
