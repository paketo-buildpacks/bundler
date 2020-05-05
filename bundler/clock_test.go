package bundler_test

import (
	"testing"
	"time"

	"github.com/paketo-community/bundler/bundler"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testClock(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("Now", func() {
		it("returns the value from the given Now function", func() {
			now := time.Now()

			clock := bundler.NewClock(func() time.Time {
				return now
			})

			Expect(clock.Now()).To(Equal(now))
		})
	})
}
