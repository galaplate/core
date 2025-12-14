package supports

func MapPostgres(connecTionType string) string {
	switch connecTionType {
	case "postgres", "postgresql", "psql":
		return "postgres"
	default:
		return connecTionType
	}
}
