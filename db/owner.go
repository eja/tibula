// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"fmt"
)

// Owners retrieves a list of user IDs who are owners of the specified module ID and have certain group associations.
// The function recursively checks group memberships up to a certain depth.
func (session *TypeSession) Owners(ownerId int64, moduleId int64) (result []int64) {
	const maxDepth = 10
	var groupOwners []int64

	ejaGroups := session.ModuleGetIdByName("ejaGroups")
	ejaUsers := session.ModuleGetIdByName("ejaUsers")
	ejaModules := session.ModuleGetIdByName("ejaModules")

	groupOwners, err := session.IncludeList(`
		SELECT dstFieldId
		FROM ejaLinks
		WHERE srcModuleId = ? AND dstModuleId = ? AND srcFieldId IN (
			SELECT srcFieldId
			FROM ejaLinks
			WHERE srcModuleId = ? AND dstModuleId = ? AND dstFieldId = ? AND srcFieldId IN (
				SELECT dstFieldId
				FROM ejaLinks
				WHERE srcModuleId = ? AND srcFieldId = ? AND dstModuleId = ?
			)
		)
	`, ejaGroups, ejaUsers,
		ejaGroups, ejaUsers, ownerId,
		ejaModules, moduleId, ejaGroups)
	if err != nil {
		return
	}

	deep := maxDepth
	userOwners := []int64{ownerId}
	for deep > 0 {
		deep--
		csv := session.NumbersToCsv(userOwners)
		users, err := session.IncludeList(fmt.Sprintf("SELECT ejaId FROM ejaUsers WHERE ejaOwner IN (%s) AND ejaId NOT IN (%s)", csv, csv))
		if err != nil {
			return
		}
		if users != nil {
			group := make(map[int64]struct{})
			for _, u := range users {
				group[u] = struct{}{}
			}
			for _, u := range userOwners {
				group[u] = struct{}{}
			}
			userOwners = make([]int64, 0, len(group))
			for u := range group {
				userOwners = append(userOwners, u)
			}
		} else {
			deep = 0
		}
	}
	group := make(map[int64]struct{})
	for _, u := range userOwners {
		group[u] = struct{}{}
	}
	for _, u := range groupOwners {
		group[u] = struct{}{}
	}

	for u := range group {
		result = append(result, u)
	}

	return
}

// OwnersCsv retrieves a comma-separated string of user IDs who are owners of the specified module ID and have certain group associations.
func (session *TypeSession) OwnersCsv(ownerId int64, moduleId int64) string {
	return session.NumbersToCsv(session.Owners(ownerId, moduleId))
}
