package authz

// moduleResources are resources that get admin/writer/reader roles.
var moduleResources = []Resource{ResourceFinance, ResourceSacrament}

// policies returns the static role -> (obj, act) rules ("p" lines).
// User -> role -> scope ("g" lines) come from the Membership table.
func policies() [][]string {
	p := [][]string{
		{RoleSuperadmin, string(ResourceDiocese), string(ActionRead)},
		{RoleSuperadmin, string(ResourceDiocese), string(ActionWrite)},
		{RoleSuperadmin, string(ResourceParish), string(ActionRead)},
		{RoleSuperadmin, string(ResourceParish), string(ActionWrite)},
	}

	for _, res := range moduleResources {
		obj := string(res)
		p = append(p,
			[]string{ModuleRole(res, LevelAdmin), obj, string(ActionRead)},
			[]string{ModuleRole(res, LevelAdmin), obj, string(ActionWrite)},
			[]string{ModuleRole(res, LevelWriter), obj, string(ActionRead)},
			[]string{ModuleRole(res, LevelWriter), obj, string(ActionWrite)},
			[]string{ModuleRole(res, LevelReader), obj, string(ActionRead)},
		)
	}
	return p
}
