// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>
package db

func (session *TypeSession) Owners(ownerId int64, moduleId int64) (result []int64) {
	uniqueOwners := make(map[int64]struct{})

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

	if err == nil {
		for _, id := range groupOwners {
			uniqueOwners[id] = struct{}{}
		}
	}

	userHierarchyQuery := `
		WITH RECURSIVE user_tree AS (
			SELECT ejaId FROM ejaUsers WHERE ejaOwner = ? AND ejaId != ejaOwner
			UNION
			SELECT u.ejaId 
			FROM ejaUsers u
			INNER JOIN user_tree ut ON u.ejaOwner = ut.ejaId
			WHERE u.ejaId != u.ejaOwner
		)
		SELECT ejaId FROM user_tree
	`

	userOwners, err := session.IncludeList(userHierarchyQuery, ownerId)
	if err == nil {
		for _, id := range userOwners {
			uniqueOwners[id] = struct{}{}
		}
	}

	uniqueOwners[ownerId] = struct{}{}

	result = make([]int64, 0, len(uniqueOwners))
	for id := range uniqueOwners {
		result = append(result, id)
	}

	return
}

func (session *TypeSession) OwnersCsv(ownerId int64, moduleId int64) string {
	return session.NumbersToCsv(session.Owners(ownerId, moduleId))
}
