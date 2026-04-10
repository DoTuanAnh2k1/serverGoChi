package leader

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

func TestMain(m *testing.M) {
	testutil.InitTestLogger()
	os.Exit(m.Run())
}

// ── nextDailyOccurrence ───────────────────────────────────────────────────────

func TestNextDailyOccurrence_AlwaysPositive(t *testing.T) {
	for hour := 0; hour <= 23; hour++ {
		d := nextDailyOccurrence(hour)
		if d <= 0 {
			t.Errorf("nextDailyOccurrence(%d) = %v, want > 0", hour, d)
		}
		if d > 24*time.Hour {
			t.Errorf("nextDailyOccurrence(%d) = %v, want <= 24h", hour, d)
		}
	}
}

func TestNextDailyOccurrence_PastHourScheduledForNextDay(t *testing.T) {
	now := time.Now()
	if now.Hour() == 0 {
		t.Skip("skip at midnight to avoid edge case")
	}
	pastHour := now.Hour() - 1
	d := nextDailyOccurrence(pastHour)

	nextTime := time.Now().Add(d)
	tomorrow := now.AddDate(0, 0, 1)
	if nextTime.Day() != tomorrow.Day() || nextTime.Month() != tomorrow.Month() {
		t.Errorf("nextDailyOccurrence(%d) lands on %s, want tomorrow (%s)",
			pastHour, nextTime.Format("2006-01-02"), tomorrow.Format("2006-01-02"))
	}
}

func TestNextDailyOccurrence_FutureHourIsToday(t *testing.T) {
	now := time.Now()
	if now.Hour() >= 22 {
		t.Skip("skip near midnight to avoid edge case")
	}
	futureHour := now.Hour() + 2
	d := nextDailyOccurrence(futureHour)
	if d > 3*time.Hour {
		t.Errorf("nextDailyOccurrence(%d) = %v, want <= 3h (today)", futureHour, d)
	}
	if d <= 0 {
		t.Errorf("nextDailyOccurrence(%d) = %v, want > 0", futureHour, d)
	}
}

// ── nextMonthlyCleanup ────────────────────────────────────────────────────────

func TestNextMonthlyCleanup_AlwaysPositive(t *testing.T) {
	d := nextMonthlyCleanup()
	if d <= 0 {
		t.Errorf("nextMonthlyCleanup() = %v, want > 0", d)
	}
	if d > 32*24*time.Hour {
		t.Errorf("nextMonthlyCleanup() = %v, want <= 32 days", d)
	}
}

func TestNextMonthlyCleanup_LandsOnFirstOfNextMonth(t *testing.T) {
	d := nextMonthlyCleanup()
	next := time.Now().Add(d)
	if next.Day() != 1 {
		t.Errorf("nextMonthlyCleanup should land on day 1, got day %d", next.Day())
	}
	if next.Hour() != 0 || next.Minute() != 0 {
		t.Errorf("nextMonthlyCleanup should land at 00:00, got %02d:%02d", next.Hour(), next.Minute())
	}
}

// ── RunTasks ─────────────────────────────────────────────────────────────────

func TestRunTasks_StopsWhenContextCancelled(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetDailyOperationHistoryFn: func(_ time.Time) ([]db_models.CliOperationHistory, error) {
			return nil, nil
		},
	})

	cfg := config_models.LeaderConfig{
		CSVExportHour: (time.Now().Hour() + 2) % 24,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		RunTasks(ctx, cfg)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("RunTasks did not stop within 3s after context cancellation")
	}
}

func TestRunTasks_ExportsImmediatelyOnStart(t *testing.T) {
	called := make(chan struct{}, 1)
	store.SetSingleton(&testutil.MockStore{
		GetDailyOperationHistoryFn: func(_ time.Time) ([]db_models.CliOperationHistory, error) {
			select {
			case called <- struct{}{}:
			default:
			}
			return nil, nil
		},
	})

	cfg := config_models.LeaderConfig{
		CSVExportHour: (time.Now().Hour() + 2) % 24,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go RunTasks(ctx, cfg)

	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("export not triggered within 2s of RunTasks starting")
	}
}

// ── runHistoryCleanup / DeleteOldHistory ──────────────────────────────────────

func TestRunHistoryCleanup_CallsDeleteWithCorrectCutoff(t *testing.T) {
	var gotCutoff time.Time
	store.SetSingleton(&testutil.MockStore{
		DeleteHistoryBeforeFn: func(cutoff time.Time) (int64, error) {
			gotCutoff = cutoff
			return 5, nil
		},
	})

	now := time.Date(2026, 4, 9, 12, 0, 0, 0, time.Local)
	// cutoff must be 2026-03-01 00:00:00
	wantCutoff := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)

	runHistoryCleanupAt(now)

	if !gotCutoff.Equal(wantCutoff) {
		t.Errorf("cutoff: got %s, want %s", gotCutoff.Format("2006-01-02"), wantCutoff.Format("2006-01-02"))
	}
}

func TestRunHistoryCleanup_JanuaryCutsToDecemberPreviousYear(t *testing.T) {
	var gotCutoff time.Time
	store.SetSingleton(&testutil.MockStore{
		DeleteHistoryBeforeFn: func(cutoff time.Time) (int64, error) {
			gotCutoff = cutoff
			return 0, nil
		},
	})

	// Jan 2026 → cutoff = 2025-12-01
	now := time.Date(2026, 1, 15, 0, 0, 0, 0, time.Local)
	wantCutoff := time.Date(2025, 12, 1, 0, 0, 0, 0, time.Local)

	runHistoryCleanupAt(now)

	if !gotCutoff.Equal(wantCutoff) {
		t.Errorf("cutoff: got %s, want %s", gotCutoff.Format("2006-01-02"), wantCutoff.Format("2006-01-02"))
	}
}

func TestRunHistoryCleanup_DBError(t *testing.T) {
	dbErr := errors.New("connection refused")
	store.SetSingleton(&testutil.MockStore{
		DeleteHistoryBeforeFn: func(_ time.Time) (int64, error) {
			return 0, dbErr
		},
	})

	// Must not panic even if DB errors
	runHistoryCleanupAt(time.Now())
}
