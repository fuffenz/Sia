package hostdb

// hostentry.go defines helper functions around the modules.HostEntry type,
// including functions for fetching settings from the host, determining the
// host weight, and adding a new (unkown) host entry into the host database
// (which requires steps like fetching the settings from the host.

import (
	"math/big"

	"github.com/NebulousLabs/Sia/modules"
	"github.com/NebulousLabs/Sia/types"
)

var (
	// Because most weights would otherwise be fractional, we set the base
	// weight to 10^80 to give ourselves lots of precision when determing the
	// weight of an entry.
	baseWeight = types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(120), nil))
)

// priceWeight returns the weight of an entry according to the price of the
// entry. The current equation is:
//		(1 / price^3)
func (hdb *HostDB) priceWeight(entry modules.HostEntry) (weight types.Currency) {
	// Prevent a divide by zero error by making sure the price is at least one.
	price := entry.Price
	if price.Cmp(types.NewCurrency64(0)) <= 0 {
		price = types.NewCurrency64(1)
	}

	// Divide the base weight by the cube of the price.
	return baseWeight.Div(price).Div(price).Div(price)
}

// insert adds a host entry to the state. The host will be inserted into the
// set of all hosts, and if it is online and responding to requests it will be
// put into the list of active hosts.
func (hdb *HostDB) insertHost(entry modules.HostEntry) {
	// Add the host to allHosts.
	entry.Reliability = types.NewCurrency64(5)
	hdb.allHosts[entry.IPAddress] = &entry

	// TODO: Add sybil attack prevention mechanism.

	go hdb.threadedProbeHost(&entry)
}

// Remove deletes an entry from the hostdb.
func (hdb *HostDB) removeHost(addr modules.NetAddress) error {
	delete(hdb.allHosts, addr)

	// See if the node is in the set of active hosts.
	node, exists := hdb.activeHosts[addr]
	if exists {
		delete(hdb.activeHosts, addr)
		node.removeNode()
	}

	return nil
}
