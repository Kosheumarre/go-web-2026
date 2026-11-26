package store

import (
	"slices"
	"sync"
)

var (
	muInscriptions sync.Mutex
	inscriptions   = map[int][]string{}
)

func Register(tournamentID int, pseudo string) bool {
	muInscriptions.Lock()
	defer muInscriptions.Unlock()
	if slices.Contains(inscriptions[tournamentID], pseudo) {
		return false
	}
	inscriptions[tournamentID] = append(inscriptions[tournamentID], pseudo)
	return true
}

func Registered(tournamentID int) []string {
	muInscriptions.Lock()
	defer muInscriptions.Unlock()
	return slices.Clone(inscriptions[tournamentID])
}