package fluxmonitorv2_test

import (
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink/core/internal/gethwrappers/generated/flux_aggregator_wrapper"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/services/fluxmonitorv2"
	"github.com/stretchr/testify/assert"
)

var (
	defaultDuration = 200 * time.Millisecond
)

type tickChecker struct {
	pollTicked  bool
	idleTicked  bool
	roundTicked bool
}

// watchTicks watches the PollManager for ticks for the waitDuration
func watchTicks(t *testing.T, pm *fluxmonitorv2.PollManager, waitDuration time.Duration) tickChecker {
	ticks := tickChecker{
		pollTicked:  false,
		idleTicked:  false,
		roundTicked: false,
	}

	waitCh := time.After(waitDuration)
	for {
		select {
		case <-pm.PollTickerTicks():
			ticks.pollTicked = true
		case <-pm.IdleTimerTicks():
			ticks.idleTicked = true
		case <-pm.RoundTimerTicks():
			ticks.roundTicked = true
		case <-waitCh:
			waitCh = nil
		}

		if waitCh == nil {
			break
		}
	}

	return ticks
}

func TestPollManager_PollTicker(t *testing.T) {
	t.Parallel()

	pm := fluxmonitorv2.NewPollManager(fluxmonitorv2.PollManagerConfig{
		PollTickerInterval: defaultDuration,
		PollTickerDisabled: false,
		IdleTimerPeriod:    defaultDuration,
		IdleTimerDisabled:  true,
	}, logger.Default)

	pm.Start(false, flux_aggregator_wrapper.OracleRoundState{})
	t.Cleanup(pm.Stop)

	ticks := watchTicks(t, pm, 2*time.Second)

	assert.True(t, ticks.pollTicked)
	assert.False(t, ticks.idleTicked)
	assert.False(t, ticks.roundTicked)
}

func TestPollManager_IdleTimer(t *testing.T) {
	t.Parallel()

	pm := fluxmonitorv2.NewPollManager(fluxmonitorv2.PollManagerConfig{
		PollTickerInterval: 100 * time.Millisecond,
		PollTickerDisabled: true,
		IdleTimerPeriod:    1 * time.Second,
		IdleTimerDisabled:  false,
	}, logger.Default)

	pm.Start(false, flux_aggregator_wrapper.OracleRoundState{
		StartedAt: uint64(time.Now().Unix()),
	})
	t.Cleanup(pm.Stop)

	ticks := watchTicks(t, pm, 2*time.Second)

	assert.False(t, ticks.pollTicked)
	assert.True(t, ticks.idleTicked)
	assert.False(t, ticks.roundTicked)
}

func TestPollManager_RoundTimer(t *testing.T) {
	t.Parallel()

	pm := fluxmonitorv2.NewPollManager(fluxmonitorv2.PollManagerConfig{
		PollTickerInterval: defaultDuration,
		PollTickerDisabled: true,
		IdleTimerPeriod:    defaultDuration,
		IdleTimerDisabled:  true,
	}, logger.Default)

	pm.Start(false, flux_aggregator_wrapper.OracleRoundState{
		StartedAt: uint64(time.Now().Unix()),
		Timeout:   1, // in seconds
	})
	t.Cleanup(pm.Stop)

	ticks := watchTicks(t, pm, 2*time.Second)

	assert.False(t, ticks.pollTicked)
	assert.False(t, ticks.idleTicked)
	assert.True(t, ticks.roundTicked)
}

func TestPollManager_HibernationOnStartThenAwaken(t *testing.T) {
	t.Parallel()

	pm := fluxmonitorv2.NewPollManager(fluxmonitorv2.PollManagerConfig{
		PollTickerInterval: defaultDuration,
		PollTickerDisabled: false,
		IdleTimerPeriod:    1 * time.Second, // Setting this too low will cause the idle timer to fire before the assert
		IdleTimerDisabled:  false,
	}, logger.Default)

	pm.Start(true, flux_aggregator_wrapper.OracleRoundState{
		StartedAt: uint64(time.Now().Unix()),
		Timeout:   1, // 1 second timeout
	})
	t.Cleanup(pm.Stop)

	ticks := watchTicks(t, pm, 2*time.Second)

	assert.False(t, ticks.pollTicked)
	assert.False(t, ticks.idleTicked)
	assert.False(t, ticks.roundTicked)

	pm.Awaken(flux_aggregator_wrapper.OracleRoundState{
		StartedAt: uint64(time.Now().Unix()),
		Timeout:   1,
	})

	ticks = watchTicks(t, pm, 2*time.Second)

	assert.True(t, ticks.pollTicked)
	assert.True(t, ticks.idleTicked)
	assert.True(t, ticks.roundTicked)
}

func TestPollManager_AwakeOnStartThenHibernate(t *testing.T) {
	t.Parallel()

	pm := fluxmonitorv2.NewPollManager(fluxmonitorv2.PollManagerConfig{
		IsHibernating:      false,
		PollTickerInterval: defaultDuration,
		PollTickerDisabled: false,
		IdleTimerPeriod:    1 * time.Second, // Setting this too low will cause the idle timer to fire before the assert
		IdleTimerDisabled:  false,
	}, logger.Default)

	pm.Start(false, flux_aggregator_wrapper.OracleRoundState{
		StartedAt: uint64(time.Now().Unix()),
		Timeout:   1,
	})
	t.Cleanup(pm.Stop)

	ticks := watchTicks(t, pm, 2*time.Second)

	assert.True(t, ticks.pollTicked)
	assert.True(t, ticks.idleTicked)
	assert.True(t, ticks.roundTicked)

	pm.Hibernate()

	ticks = watchTicks(t, pm, 2*time.Second)

	assert.False(t, ticks.pollTicked)
	assert.False(t, ticks.idleTicked)
	assert.False(t, ticks.roundTicked)
}

func TestPollManager_ShouldPerformInitialPoll(t *testing.T) {
	testCases := []struct {
		name               string
		pollTickerDisabled bool
		idleTimerDisabled  bool
		isHibernating      bool
		want               bool
	}{
		{
			name:               "perform poll - all enabled",
			pollTickerDisabled: false,
			idleTimerDisabled:  false,
			isHibernating:      false,
			want:               true,
		},
		{
			name:               "don't perform poll - hibernating",
			pollTickerDisabled: false,
			idleTimerDisabled:  false,
			isHibernating:      true,
			want:               false,
		},
		{
			name:               "perform poll - only pollTickerDisabled",
			pollTickerDisabled: true,
			idleTimerDisabled:  false,
			isHibernating:      false,
			want:               true,
		},
		{
			name:               "perform poll - only idleTimerDisabled",
			pollTickerDisabled: false,
			idleTimerDisabled:  true,
			isHibernating:      false,
			want:               true,
		},
		{
			name:               "don't perform poll - idleTimerDisabled and pollTimerDisabled",
			pollTickerDisabled: true,
			idleTimerDisabled:  true,
			isHibernating:      false,
			want:               false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pm := fluxmonitorv2.NewPollManager(fluxmonitorv2.PollManagerConfig{
				IsHibernating:      tc.isHibernating,
				PollTickerInterval: defaultDuration,
				PollTickerDisabled: tc.pollTickerDisabled,
				IdleTimerPeriod:    defaultDuration,
				IdleTimerDisabled:  tc.idleTimerDisabled,
			}, logger.Default)

			assert.Equal(t, tc.want, pm.ShouldPerformInitialPoll())
		})

	}
}
