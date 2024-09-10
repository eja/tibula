// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

// TypeGroup represents a modular structure containing information about groups.
type TypeGroup struct {
	Name        string              `json:"name"`
	Type        string              `json:"type`
	Shares      []string            `json:"shares,omitempty"`
	Permissions map[string][]string `json:"permissions,omitempty"`
}

// UserGroupList retrieves the list of group IDs associated with a user.
func (session *TypeSession) UserGroupList(userId int64) []int64 {
	response, err := session.IncludeList("SELECT srcFieldId FROM ejaLinks WHERE srcModuleId=? AND dstModuleId=? AND dstFieldId=?", session.ModuleGetIdByName("ejaGroups"), session.ModuleGetIdByName("ejaUsers"), userId)
	if err != nil || len(response) == 0 {
		return []int64{0}
	}
	return response
}

// UserGroupCsv returns a comma-separated string of group IDs associated with a user.
func (session *TypeSession) UserGroupCsv(userId int64) string {
	return session.NumbersToCsv(session.UserGroupList(userId))
}
