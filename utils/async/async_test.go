package async

import (
	"fmt"
	"testing"

	"github.com/michael-reichenauer/gmc/utils/one"
)

func TestRun(t *testing.T) {
	one.Run(func() {
		t.Log("test")

		as := Run(func() (int, error) { return call(2) }).
			Then(func(r int) { t.Logf("Result1 %d", r) }).
			Catch(func(e error) { t.Logf("Error: %s", e) }).
			Finally(func() { t.Log("Finally done") })

		ThenRun(as, func(r int) (string, error) { return otherCall(r) }).
			Then(func(r string) { t.Logf("other result: %q", r) }).
			Catch(func(e error) { t.Logf("other Error: %v", e) }).
			Finally(func() {
				t.Logf("Other finally done")
				one.Close()
			})
	})
}

func call(count int) (int, error) {
	if count > 4 {
		return 0, fmt.Errorf("value to large")
	}

	return count * 2, nil
}

func otherCall(count int) (string, error) {
	if count > 4 {
		return "", fmt.Errorf("other value to large")
	}

	return fmt.Sprintf("other value '%d'", count*2), nil
}
