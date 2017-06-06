package handlers

import (
	"fmt"
	"policy-server/models"
	"policy-server/uaa_client"
)

type QuotaGuard struct {
	Store       store
	MaxPolicies int
}

func (g *QuotaGuard) CheckAccess(policies []models.Policy, userToken uaa_client.CheckTokenResponse) (bool, error) {
	for _, scope := range userToken.Scope {
		if scope == "network.admin" {
			return true, nil
		}
	}
	appGuids := uniqueAppGUIDs(policies)
	toAddSourceCounts := sourceCounts(policies, appGuids)
	sourcePolicies, err := g.Store.ByGuids(appGuids, []string{})
	if err != nil {
		return false, fmt.Errorf("getting policies: %s", err)
	}
	currentAppCounts := sourceCounts(sourcePolicies, appGuids)
	for _, appGuid := range appGuids {
		if currentAppCounts[appGuid]+toAddSourceCounts[appGuid] > g.MaxPolicies {
			return false, nil
		}
	}
	return true, nil
}

func sourceCounts(policies []models.Policy, knownAppGuids []string) map[string]int {
	var set = make(map[string]int)
	for _, appGuid := range knownAppGuids {
		set[appGuid] = 0
	}
	for _, policy := range policies {
		set[policy.Source.ID] = set[policy.Source.ID] + 1
	}
	return set
}
