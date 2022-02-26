package person

type Student struct {
	name  string
	score float32
}

func (s *Student) GetName() string {
	return s.name
}

func (s *Student) SetName(newName string) {
	s.name = newName
}
