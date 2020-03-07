package collector

// Enhanced structure to store all enhanced metrics
var Enhanced = []Metric{
	{
		Kind:        "Enhanced",
		Group:       "General",
		Name:        "engine",
		Units:       "",
		Description: "The database engine for the DB instance.",
	},
	{
		Kind:        "Enhanced",
		Group:       "General",
		Name:        "instanceID",
		Units:       "",
		Description: "The DB instance identifier.",
	},
}
