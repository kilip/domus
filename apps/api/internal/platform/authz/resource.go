package authz

type Resource string
type Action string

const (
	ResourceDiocese   Resource = "dioceses"
	ResourceParish    Resource = "parishes"
	ResourceFinance   Resource = "finance"
	ResourceSacrament Resource = "sacrament"
)

const (
	ActionRead  Action = "read"
	ActionWrite Action = "write"
)

type Level string

const (
	LevelAdmin  Level = "admin"
	LevelWriter Level = "writer"
	LevelReader Level = "reader"
)

const RoleSuperadmin = "superadmin"

// ModuleRole builds a per-module role name, e.g. "finance-writer".
func ModuleRole(res Resource, level Level) string {
	return string(res) + "-" + string(level)
}
