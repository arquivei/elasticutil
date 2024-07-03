package v7

// Sorters represents a list of Sorter.
type Sorters struct {
	Sorters []Sorter
}

// Sorter represents a Elasticsearch´s sort structure.
type Sorter struct {
	Field     string
	Ascending bool
}

func (s Sorter) String() string {
	direction := "asc"
	if !s.Ascending {
		direction = "desc"
	}
	return s.Field + ":" + direction
}

func (ss Sorters) Strings() []string {
	response := make([]string, 0, len(ss.Sorters))
	for _, sorter := range ss.Sorters {
		response = append(response, sorter.String())
	}
	return response
}
