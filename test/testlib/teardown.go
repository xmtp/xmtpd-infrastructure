package testlib

import (
	"github.com/stretchr/testify/require"
	"testing"
)

/** Lists of the teardown and diagnostic teardown funcs */
var teardownLists = make(map[string][]func())

// AddTeardown
/**
 * add a teardown function to the named list - for deferred execution.
 *
 * The teardown functions are called in reverse order of insertion, by a call to Teardown(name).
 *
 * The typical idiom is:
 * <pre>
 *   testlib.AddTeardown("MLS", func() { ...})
 *   // possibly more testlib.AddTeardown("MLS", func() { ... })
 *   defer testlib.Teardown("MLS")
 * <pre>
 */
func AddTeardown(name string, teardownFunc func()) {
	teardownLists[name] = append(teardownLists[name], teardownFunc)
}

// Teardown
/**
 * Call the stored teardown functions in the named list, in the correct order (last-in-first-out)
 * The typical use of Teardown is with a deferred call:
 * defer testlib.Teardown("MLS")
 */
func Teardown(name string) {
	// ensure both list and diagnostic list are removed.
	defer func() { delete(teardownLists, name) }()

	list := teardownLists[name]

	for x := len(list) - 1; x >= 0; x-- {
		list[x]()
	}
}

// VerifyTeardown
/**
* Verify all teardownLists have been executed already; and throw a `require` if not.
* Can be used to verify correct coding of a test that uses teardown - and to ensure eventual release of resources.
*
* NOTE: while the funcs are called in the correct order for each list,
* there can be NO guarantee that the lists are iterated in the correct order.
*
* This function MUST NOT be used as a replacement for calling teardown() at the correct point in the code.
 */
func VerifyTeardown(t *testing.T) {
	// ensure all funcs in all lists are released
	defer func() { teardownLists = make(map[string][]func()) }()

	// release all remaining resources - this is a "best effort" as the order of iterating the map is arbitrary
	uncleared := make([]string, 0)

	// make a "best-effort" at releasing all remaining resources
	for name, list := range teardownLists {
		uncleared = append(uncleared, name)

		for x := len(list) - 1; x >= 0; x-- {
			list[x]()
		}
	}

	require.Equal(t, 0, len(uncleared), "Error - %d teardownLists were left uncleared: %s", len(uncleared), uncleared)
	t.Log("Waiting for all logging collectors to finish")
	appLogCollectorsWg.Wait()
}
