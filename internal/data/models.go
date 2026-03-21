package data

type Models struct {
	Persons         PersonModel
}

func (m Models) NewModels() Models {
	return Models{
		Persons: PersonModel{},
	}
}