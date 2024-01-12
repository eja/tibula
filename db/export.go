// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

// ModuleExport exports a module.
func ModuleExport(moduleId int64, data bool) (module TypeModule, err error) {
	var row TypeRow
	var rows TypeRows
	moduleName := ModuleGetNameById(moduleId)
	module.Name = moduleName
	row, err = Row("SELECT a.searchLimit, a.sqlCreated, a.power, a.sortList, (SELECT x.name FROM ejaModules AS x WHERE x.ejaId=a.parentId) AS parentName FROM ejaModules AS a WHERE ejaId=?", moduleId)
	if err != nil {
		return
	}
	module.Module = TypeModuleModule{
		ParentName:  row["parentName"],
		Power:       Number(row["power"]),
		SearchLimit: Number(row["searchLimit"]),
		SqlCreated:  Number(row["sqlCreated"]),
		SortList:    row["sortList"],
	}

	rows, err = Rows("SELECT translate, powerEdit, name, type, powerList, powerSearch, value FROM ejaFields WHERE ejaModuleId=?", moduleId)
	if err != nil {
		return
	}
	for _, row := range rows {
		module.Field = append(module.Field, TypeModuleField{
			Name:        row["name"],
			Value:       row["value"],
			PowerSearch: Number(row["powerSearch"]),
			PowerList:   Number(row["powerList"]),
			PowerEdit:   Number(row["powerEdit"]),
			Type:        row["type"],
			Translate:   Number(row["translate"]),
		})
	}

	rows, err = Rows(`
		SELECT ejaLanguage, word, translation, (SELECT ejaModules.name FROM ejaModules WHERE ejaModules.ejaId=ejaModuleId) AS ejaModuleName 
		FROM ejaTranslations 
		WHERE ejaModuleId=? OR word='?'
		`, moduleId, moduleName)
	if err != nil {
		return
	}
	for _, row := range rows {
		module.Translation = append(module.Translation, TypeModuleTranslation{
			EjaLanguage:   row["ejaLangauge"],
			EjaModuleName: row["ejaModuleName"],
			Word:          row["word"],
			Translation:   row["translation"],
		})
	}

	module.Command = []string{}
	rows, err = Rows("SELECT name from ejaCommands WHERE ejaId IN (SELECT ejaCommandId FROM ejaPermissions WHERE ejaModuleId=?)", moduleId)
	if err != nil {
		return
	}
	for _, row := range rows {
		module.Command = append(module.Command, row["name"])
	}

	if data {
		rows, err = Rows("SELECT * FROM " + moduleName)
		if err != nil {
			return
		}
		for idx, row := range rows {
			module.Data = append(module.Data, make(map[string]interface{}))
			for key, val := range row {
				if key != "ejaId" && key != "ejaOwner" && key != "ejaLog" {
					module.Data[idx][key] = val
				}
			}
		}
	}

	return
}
