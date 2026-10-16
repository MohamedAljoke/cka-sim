package vertex

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"testing"

	"google.golang.org/genai"
)

func TestBatchesKeepTheOrderAndStayUnderBothLimits(t *testing.T) {
	got := batches([]string{"aa", "bb", "cc", "dddddd", "e"}, 2, 6)

	want := [][]string{{"aa", "bb"}, {"cc"}, {"dddddd"}, {"e"}}
	if !slices.EqualFunc(got, want, slices.Equal) {
		t.Errorf("batches = %q, want %q", got, want)
	}
}

func TestBatchesPutsATooLongTextAloneRatherThanDroppingIt(t *testing.T) {
	got := batches([]string{"0123456789"}, 50, 4)

	if len(got) != 1 || got[0][0] != "0123456789" {
		t.Errorf("batches = %q", got)
	}
}

func TestRetryingWaitsOutRateLimitsOnly(t *testing.T) {
	calls := 0
	err := retrying(context.Background(), 5, 0, func() error {
		calls++
		if calls < 3 {
			return genai.APIError{Code: http.StatusTooManyRequests}
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Errorf("err = %v after %d calls", err, calls)
	}

	calls = 0
	err = retrying(context.Background(), 5, 0, func() error {
		calls++
		return genai.APIError{Code: http.StatusBadRequest}
	})
	if apiErr, ok := errors.AsType[genai.APIError](err); !ok || apiErr.Code != http.StatusBadRequest || calls != 1 {
		t.Errorf("err = %v after %d calls", err, calls)
	}
}

func TestRetryingGivesUpAfterItsTries(t *testing.T) {
	calls := 0
	err := retrying(context.Background(), 3, 0, func() error {
		calls++
		return genai.APIError{Code: http.StatusTooManyRequests}
	})

	if err == nil || calls != 3 {
		t.Errorf("err = %v after %d calls", err, calls)
	}
}
