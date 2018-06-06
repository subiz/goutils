package migratev3

var accidm = map[string]string {

}

var agentidm = map[string]string {
}

func IsCollided(v3accid string) bool {
	_, ok := accidm[v3accid]
	return ok
}

func MapToAccountV4(v3accid string) string {
	return accidm[v3accid]
}

func MapToAgentV3(v3agid string) string {
	return agentidm[v3agid]
}
