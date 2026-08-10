package trainingcenter

type Repository interface {
	List() []TrainingCenter
	Exists(id string) bool
}