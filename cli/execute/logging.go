package execute

import (
	"github.com/ThisWillGoWell/saw/blade"
	sawConfig "github.com/ThisWillGoWell/saw/config"
	"os"
	"os/signal"
)

func  FollowLogs(logGroup, startTime, endTime, filter string ) error {
	awsSawConfig := &sawConfig.AWSConfiguration{
		Region:  "us-east-2",
	}

	watchConfig := &sawConfig.Configuration{
		Group:     logGroup,
		Prefix:     "",
		Start:      startTime,
		End:        endTime,
		Filter:     filter,
		Streams:    nil,
		Descending: false,
		OrderBy:    "",
	}

	outputConfig := &sawConfig.OutputConfiguration{
		Raw:       false,
		Pretty:    true,
		Expand:    false,
		Invert:    false,
		RawString: false,
		NoColor:   false,
	}
	b := blade.NewBlade(watchConfig, awsSawConfig, outputConfig)

	cancel := make(chan interface{}, 1)
	go b.StreamEventsWithCancel(cancel)

	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, os.Interrupt)
	<- interruptSignal
	close(cancel)


	return nil
}

