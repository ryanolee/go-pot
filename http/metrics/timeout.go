package metrics

import (
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/ryanolee/ryan-pot/http/gossip/action"
	"go.uber.org/zap"
)

const (
	// If we have a timeout that is really long we can immidiately commit it to the cold cache
	instantCommitThreshold = time.Second * 180

	// The upper bound for timeouts before we hang forever
	upperTimeoutBound = time.Second * 60

	// The smallest timeout we will ever give
	lowerTimeoutBound = time.Second * 1

	// The increment we will increase timeouts by for requests with timeouts larger than 30 seconds
	timeoutIncrement = time.Second * 10

	// The increment we will increase timeouts by for requests with timeouts smaller than 30 seconds
	timeoutSubThirtyIncrement = time.Second * 5

	// The increment we will increase timeouts by for requests with timeouts smaller than 10 seconds
	timeoutSubTenIncrement = time.Second * 2

	// Grace requests is the number of requests we will allow before beginning to try and timeout the IP
	GraceRequests = 3
	GraceTimeout  = time.Millisecond * 100

	// If we know the IP will hang on forever grant them what they want!
	longestTimeout = time.Hour * 96

	sampleSize      = 3
	sampleDeviation = time.Second * 1
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

		actionDispatcher action.BroadcastActionDispatcher
	}

	// Timeout for an IP address we have been able to work out who's timeout is
	CommittedTimeoutForIp struct {
		Timeout time.Duration
		Ip      string
	}

	TimeoutForIp struct {
		mutex                sync.RWMutex
		Requests             int
		ValidTimeouts        []time.Duration
		InvalidTimeouts      []time.Duration
		LastValidTimeout     time.Duration
		LastInvalidTimeout   time.Duration
		LastPerformedTimeout time.Duration
	}
)

func NewTimeoutForIp() *TimeoutForIp {
	return &TimeoutForIp{
		mutex:                sync.RWMutex{},
		Requests:             0,
		ValidTimeouts:        make([]time.Duration, 0),
		InvalidTimeouts:      make([]time.Duration, 0),
		LastValidTimeout:     0,
		LastInvalidTimeout:   0,
		LastPerformedTimeout: 0,
	}
}

func (t *TimeoutForIp) CalculateNextTimeout() time.Duration {

	if t.Requests < GraceRequests {
		return GraceTimeout
	}

	if t.LastPerformedTimeout < time.Second*10 {
		return t.LastPerformedTimeout + timeoutSubTenIncrement
	}

	if t.LastPerformedTimeout < time.Second*30 {
		return t.LastPerformedTimeout + timeoutSubThirtyIncrement
	}

	if t.LastPerformedTimeout < upperTimeoutBound {
		return t.LastPerformedTimeout + timeoutIncrement
	}

	return longestTimeout
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

	startingPos := len(t.InvalidTimeouts) - sampleSize
	squaredSum := float64(0)
	for i := startingPos; i < len(t.InvalidTimeouts); i++ {
		squaredSum += math.Pow(math.Abs(avg-float64(t.InvalidTimeouts[i])), 2)
	}

	return time.Duration(math.Sqrt(squaredSum / sampleSize))
}

func (t *TimeoutForIp) GetAverageTimeoutInSample() time.Duration {
	if len(t.InvalidTimeouts) < sampleSize {
		return -1
	}

	sum := float64(0)
	for i := 0; i < len(t.InvalidTimeouts); i++ {
		sum += float64(t.InvalidTimeouts[i])
	}
	return time.Duration(sum / sampleSize)
}

func (t *TimeoutForIp) RecordInvalidTimeout(timeout time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.InvalidTimeouts = append(t.InvalidTimeouts, timeout)
	t.LastInvalidTimeout = timeout

	if len(t.InvalidTimeouts) > sampleSize {
		t.InvalidTimeouts = t.InvalidTimeouts[1:]
	}
}

func (t *TimeoutForIp) RecordValidTimeout(timeout time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.ValidTimeouts = append(t.ValidTimeouts, timeout)
	t.LastValidTimeout = timeout

	if len(t.ValidTimeouts) > sampleSize {
		t.ValidTimeouts = t.ValidTimeouts[1:]
	}
}

func NewTimeoutWatcher() *TimeoutWatcher {
	return &TimeoutWatcher{
		hotCachePool:  cache.New(time.Minute*30, time.Minute),
		coldCachePool: cache.New(time.Hour*24, time.Hour),
	}
}

func (tw *TimeoutWatcher) RecordResponse(ipAddress string, timeout time.Duration, successful bool) {
	var data *TimeoutForIp
	result, ok := tw.hotCachePool.Get(ipAddress)

	if !ok {
		result = NewTimeoutForIp()
	}

	if data, ok = result.(*TimeoutForIp); !ok {
		zap.L().Sugar().Warn("Failed to cast timeout data for IP address. Resetting", "ip", ipAddress)
		data = NewTimeoutForIp()
	}

	if successful {
		data.RecordValidTimeout(timeout)
	} else {
		data.RecordInvalidTimeout(timeout)
	}

	if !successful && timeout > instantCommitThreshold {
		zap.L().Sugar().Infow("Timeout recorded higher than instant commit threshold", "ip", ipAddress, "timeout", timeout)
		tw.CommitToColdCacheWithBroadcast(ipAddress, longestTimeout)
		return
	}

	if len(data.InvalidTimeouts) < sampleSize {
		return
	}

	sd := data.GetStandardDeviation()
	if sd < 0 {
		return
	}

	if sd > sampleDeviation {
		return
	}
	zap.L().Sugar().Infow("Standard deviation is low. We have probably found the timeout! Committing to cold cache", "ip", ipAddress, "sd", sd)
	avg := data.GetAverageTimeoutInSample()
	timeoutToCommit := avg - (sd * 2)
	if timeoutToCommit < lowerTimeoutBound {
		timeoutToCommit = lowerTimeoutBound
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
	tw.actionDispatcher.Broadcast(&action.BroadcastAction{
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
		result = NewTimeoutForIp()
	}

	if data, ok = result.(*TimeoutForIp); !ok {
		zap.L().Sugar().Warn("Failed to cast timeout data for IP address. Resetting", "ip", ipAddress)
		data = NewTimeoutForIp()
	}

	tw.hotCachePool.Set(ipAddress, data, cache.DefaultExpiration)

	timeout := data.GetNextTimeout()

	return timeout
}

func (tw *TimeoutWatcher) BindActionDispatcher(dispatcher action.BroadcastActionDispatcher) {
	tw.actionDispatcher = dispatcher
}
