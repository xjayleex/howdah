package utils

import (
	"4d63.com/tz"
	"errors"
	"google.golang.org/protobuf/types/known/timestamppb"
	"howdah/internal/pkg/common/consts"
	"time"
)

type Timestamper struct {
	opts           timestampOptions
	loadedLocation *time.Location
}

func NewTimestamper (opt ...TimestampOption) *Timestamper {
	opts := defaultTimestamperOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	timestamper := &Timestamper{
		opts: opts,
	}

	loaded, err := tz.LoadLocation(opts.location)
	if err != nil {
		if opts.location == consts.DefaultTimezone {
			// Todo : panic
		} else {
			loaded, _ = tz.LoadLocation(consts.DefaultTimezone)
		}
	}

	timestamper.loadedLocation = loaded

	return timestamper
}

func (t *Timestamper) Now() *timestamppb.Timestamp {
	now := time.Now().In(t.loadedLocation)
	return timestamppb.New(now)
}

type TimestampOption interface {
	apply(*timestampOptions)
}

type timestampOptions struct {
	location string
}
var defaultTimestamperOptions = timestampOptions{
	location: consts.DefaultTimezone,
}

func WithNewLocation (location string) TimestampOption {
	return newFuncTimestampOption(func(o *timestampOptions) {
		o.location = location
	})
}


type funcTimestampOption struct {
	f func(*timestampOptions)
}

func (fto *funcTimestampOption) apply(to *timestampOptions) {
	fto.f(to)
}

func newFuncTimestampOption(f func(*timestampOptions)) *funcTimestampOption {
	return &funcTimestampOption{
		f: f,
	}
}

type RecoveryPolicy interface {
	Interval() time.Duration
	Reset()
}

type recoveryPolicy struct {
	current time.Duration
	starting time.Duration
	threshold time.Duration
}

func NewRecoveryPolicy (starting, threshold time.Duration) (RecoveryPolicy, error){
	if starting > threshold {
		return nil, errors.New("starting duration cannot be higher than threshold")
	}
	return &recoveryPolicy{
		current:   starting,
		starting:  starting,
		threshold: threshold,
	}, nil
}

func (rp *recoveryPolicy) Interval() time.Duration {
	if rp.current * 2 < rp.threshold {
		rp.current = rp.current * 2
		return rp.current
	} else {
		return rp.threshold
	}
}

func (rp *recoveryPolicy) Reset() {
	rp.current = rp.starting
}

/*
type RetryCounter struct {
	patients	int
	counter		int
}
func NewRetryCounter (patients int) *RetryCounter {
	return &RetryCounter{
		patients: patients,
		counter:  0,
	}

}

func (r *RetryCounter) ShouldPatient() bool{
	if r.counter < r.patients {
		return true
	} else {
		return false
	}
}

func (r *RetryCounter) AddOne() {
	r.counter += 1
}

func (r *RetryCounter) Reset() {
	r.counter = 0
}
*/