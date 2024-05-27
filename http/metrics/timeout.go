package metrics

import (
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/ryanolee/ryan-pot/config"
	"github.com/ryanolee/ryan-pot/http/gossip/action"
	"go.uber.org/zap"
)

type (
	// Timeout watcher aggregates timeouts for a given IP address in order to
	// Current model for working out timeout is as follows:
	//   - Begin keeping track of timeouts of and IP address the "hot cache pool" until:
	//      - We have a timeout that is longer than "instantCommitThreshold"
	//      - We have a standard deviation of timeouts that is less than "sampleDeviation"
	//   - Then we commit the timeout to the "cold cache pool" and delete the IP from the "hot cache pool"
	//   - Broadcast the committed timeout to other nodes in the cluster
	//   - Always return the known timeout for an IP address from the "cold cache pool"

	TimeoutWatcher struct {
		// Cache pool for IP addresses we are actively trying to work out timeout for
		hotCachePool *cache.Cache

		// Cache pool for IP addresses we have already worked out the IP timeouts for
		coldCachePool *cache.Cache

		actionDispatcher action.IBroadcastActionDispatcher

		opts *TimeoutWatcherOptions
	}

	TimeoutWatcherOptions struct {
		// The maximum amount of time a given IP can be hanging before we consider the IP
		// to be vulnerable to hanging forever on a request. Any ips that get past this threshold
		// will always be given the longest timeout
		instantCommitThreshold time.Duration

		// The upper bound for increasing timeouts. Once the timeout increases to reach this bound we will hang forever.
		upperTimeoutBound time.Duration

		// The smallest timeout we will ever give
		lowerTimeoutBound time.Duration

		// The increment we will increase timeouts by for requests with timeouts larger than 30 seconds
		timeoutOverThirtyIncrement time.Duration

		// The increment we will increase timeouts by for requests with timeouts smaller than 30 seconds
		timeoutSubThirtyIncrement time.Duration

		// The increment we will increase timeouts by for requests with timeouts smaller than 10 seconds
		timeoutSubTenIncrement time.Duration

		// The number of requests that are allowed before things begin slowing down
		graceRequests int

		// The timeout we will give to requests that are allowed to pass the grace period
	    graceTimeout time.Duration

		// The amount of time to wait when hanging an IP "forever"
		longestTimeout time.Duration

		// The number of samples to take to detect a timeout
		sampleSize 	int

		// How close the standards need to be on average to move the IP address into the "Endless stall" category
		sampleDeviation time.Duration
	}

	// Timeout for an IP address we have been able to work out who's timeout is
	CommittedTimeoutForIp struct {
		Timeout time.Duration
		Ip      string
	}

	// Struct representing the traffic for a given IP address
	TimeoutForIp struct {
		// Options for the timeout watcher associated with this IP Timeout
		opts              	*TimeoutWatcherOptions

		// Mutex for sync operations relating to the given IP
		mutex                sync.RWMutex

		// The number of requests that have been made by an given IP to this node
		Requests             int

		// The duration of the last N timeouts that finished successfully
		ValidTimeouts        []time.Duration

		// The duration of the last N timeouts that finished with a client timeout
		InvalidTimeouts      []time.Duration

		// The duration of the last valid timeout
		LastValidTimeout     time.Duration

		// The duration of the last invalid timeout
		LastInvalidTimeout   time.Duration

		// The duration of the last timeout that was attempted
		LastPerformedTimeout time.Duration
	}
)

func NewTimeoutForIp(opts *TimeoutWatcherOptions) *TimeoutForIp {
	return &TimeoutForIp{
		mutex:                sync.RWMutex{},
		opts:                 opts,
		Requests:             0,
		ValidTimeouts:        make([]time.Duration, 0),
		InvalidTimeouts:      make([]time.Duration, 0),
		LastValidTimeout:     0,
		LastInvalidTimeout:   0,
		LastPerformedTimeout: 0,
	}
}

func (t *TimeoutForIp) CalculateNextTimeout() time.Duration {
	if t.Requests < t.opts.graceRequests {
		return t.opts.graceTimeout
	}

	if t.LastPerformedTimeout < time.Second*10 {
		return t.LastPerformedTimeout + t.opts.timeoutSubTenIncrement
	}

	if t.LastPerformedTimeout < time.Second*30 {
		return t.LastPerformedTimeout + t.opts.timeoutSubThirtyIncrement
	}

	if t.LastPerformedTimeout < t.opts.upperTimeoutBound {
		return t.LastPerformedTimeout + t.opts.timeoutOverThirtyIncrement
	}

	return t.opts.longestTimeout
}

func (t *TimeoutForIp) GetNextTimeout() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	timeout := t.CalculateNextTimeout()
	t.Requests += 1
	t.LastPerformedTimeout = timeout
	return timeout
}

func (t *TimeoutForIp) GetStandardDeviation() time.Duration {
	avg := float64(t.GetAverageTimeoutInSample())
	if avg == -1 {
		return -1
	}

	startingPos := len(t.InvalidTimeouts) - t.opts.sampleSize
	squaredSum := float64(0)
	for i := startingPos; i < len(t.InvalidTimeouts); i++ {
		squaredSum += math.Pow(math.Abs(avg-float64(t.InvalidTimeouts[i])), 2)
	}

	return time.Duration(math.Sqrt(squaredSum / float64(t.opts.sampleSize)))
}

func (t *TimeoutForIp) GetAverageTimeoutInSample() time.Duration {
	if len(t.InvalidTimeouts) < t.opts.sampleSize {
		return -1
	}

	sum := float64(0)
	for i := 0; i < len(t.InvalidTimeouts); i++ {
		sum += float64(t.InvalidTimeouts[i])
	}
	return time.Duration(sum / float64(t.opts.sampleSize))
}

func (t *TimeoutForIp) RecordInvalidTimeout(timeout time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.InvalidTimeouts = append(t.InvalidTimeouts, timeout)
	t.LastInvalidTimeout = timeout

	if len(t.InvalidTimeouts) > t.opts.sampleSize {
		t.InvalidTimeouts = t.InvalidTimeouts[1:]
	}
}

func (t *TimeoutForIp) RecordValidTimeout(timeout time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.ValidTimeouts = append(t.ValidTimeouts, timeout)
	t.LastValidTimeout = timeout

	if len(t.ValidTimeouts) > t.opts.sampleSize {
		t.ValidTimeouts = t.ValidTimeouts[1:]
	}
}

func NewTimeoutWatcher(config *config.Config) *TimeoutWatcher {
	if !config.TimeoutWatcher.Enabled {
		return nil
	}

	twConfig := &config.TimeoutWatcher

	return &TimeoutWatcher{
		actionDispatcher: nil,
		hotCachePool:  cache.New(time.Duration(twConfig.CacheHotPoolTTL) * time.Second, time.Minute),
		coldCachePool: cache.New(time.Duration(twConfig.CacheColdPoolTTL) * time.Second, time.Hour),
		
		// Map options from config to TimeoutWatcherOptions
		opts: &TimeoutWatcherOptions{
			instantCommitThreshold:     time.Duration(twConfig.InstantCommitThreshold) * time.Millisecond,
			upperTimeoutBound:          time.Duration(twConfig.UpperTimeoutBound) * time.Millisecond,
			lowerTimeoutBound:          time.Duration(twConfig.LowerTimeoutBound) * time.Millisecond,
			timeoutOverThirtyIncrement: time.Duration(twConfig.TimeoutOverThirtyIncrement) * time.Millisecond,
			timeoutSubThirtyIncrement:  time.Duration(twConfig.TimeoutSubThirtyIncrement) * time.Millisecond,
			timeoutSubTenIncrement:     time.Duration(twConfig.TimeoutSubTenIncrement) * time.Millisecond,
			graceRequests:              twConfig.GraceRequests,
			graceTimeout:               time.Duration(twConfig.GraceTimeout) * time.Millisecond,
			longestTimeout:             time.Duration(twConfig.LongestTimeout) * time.Millisecond,
			sampleSize:                 twConfig.DetectionSampleSize,
			sampleDeviation:            time.Duration(twConfig.DetectionSampleDeviation) * time.Millisecond,
		},
	}
}

func (tw *TimeoutWatcher) SetActionDispatcher(actionDispatcher action.IBroadcastActionDispatcher) {
	tw.actionDispatcher = actionDispatcher
}

func (tw *TimeoutWatcher) RecordResponse(ipAddress string, timeout time.Duration, successful bool) {
	var data *TimeoutForIp
	result, ok := tw.hotCachePool.Get(ipAddress)

	if !ok {
		result = NewTimeoutForIp(tw.opts)
	}

	if data, ok = result.(*TimeoutForIp); !ok {
		zap.L().Sugar().Warn("Failed to cast timeout data for IP address. Resetting", "ip", ipAddress)
		data = NewTimeoutForIp(tw.opts)
	}

	if successful {
		data.RecordValidTimeout(timeout)
	} else {
		data.RecordInvalidTimeout(timeout)
	}

	if !successful && timeout > tw.opts.instantCommitThreshold {
		zap.L().Sugar().Infow("Timeout recorded higher than instant commit threshold", "ip", ipAddress, "timeout", timeout)
		tw.CommitToColdCacheWithBroadcast(ipAddress, tw.opts.longestTimeout)
		return
	}

	if len(data.InvalidTimeouts) < tw.opts.sampleSize {
		return
	}

	sd := data.GetStandardDeviation()
	if sd < 0 {
		return
	}

	if sd > tw.opts.sampleDeviation {
		return
	}
	zap.L().Sugar().Infow("Standard deviation is low. We have probably found the timeout! Committing to cold cache", "ip", ipAddress, "sd", sd)
	avg := data.GetAverageTimeoutInSample()
	timeoutToCommit := avg - (sd * 2)
	if timeoutToCommit < tw.opts.lowerTimeoutBound {
		timeoutToCommit = tw.opts.lowerTimeoutBound
	}

	zap.L().Sugar().Infow("Committed to cold cache", "ip", ipAddress, "timeout", timeoutToCommit)
	tw.CommitToColdCacheWithBroadcast(ipAddress, timeoutToCommit)
}

func (tw *TimeoutWatcher) CommitToColdCache(ipAddress string, timeout time.Duration) {
	tw.coldCachePool.Set(ipAddress, timeout, cache.DefaultExpiration)
	tw.hotCachePool.Delete(ipAddress)
}

func (tw *TimeoutWatcher) CommitToColdCacheWithBroadcast(ipAddress string, timeout time.Duration) {
	tw.CommitToColdCache(ipAddress, timeout)
	tw.BroadcastColdCacheIp(ipAddress, timeout)
}

func (tw *TimeoutWatcher) HasColdCacheTimeout(ipAddress string) bool {
	_, ok := tw.coldCachePool.Get(ipAddress)
	return ok
}

func (tw *TimeoutWatcher) BroadcastColdCacheIp(ipAddress string, timeout time.Duration) {
	if tw.actionDispatcher == nil {
		return
	}

	tw.actionDispatcher.Dispatch(&action.BroadcastAction{
		Action: "ADD_COLD_IP",
		Data:   ipAddress + "," + strconv.Itoa(int(timeout)),
	})
}

func (tw *TimeoutWatcher) GetTimeout(ipAddress string) time.Duration {
	tw.coldCachePool.Get(ipAddress)
	if timeout, ok := tw.coldCachePool.Get(ipAddress); ok {
		if timeout, ok := timeout.(time.Duration); ok {
			return timeout
		}

		zap.L().Sugar().Warn("Failed to cast timeout data for IP address. Resetting", "ip", ipAddress)
		tw.coldCachePool.Delete(ipAddress)
	}

	var data *TimeoutForIp
	result, ok := tw.hotCachePool.Get(ipAddress)

	if !ok {
		result = NewTimeoutForIp(tw.opts)
	}

	if data, ok = result.(*TimeoutForIp); !ok {
		zap.L().Sugar().Warn("Failed to cast timeout data for IP address. Resetting", "ip", ipAddress)
		data = NewTimeoutForIp(tw.opts)
	}

	tw.hotCachePool.Set(ipAddress, data, cache.DefaultExpiration)

	timeout := data.GetNextTimeout()

	return timeout
}

