package model

const (
	QuizStatusOpen     = 1
	QuizStatusFinished = 2
)

type QuizTab struct {
	ID          int64 `gorm:"primarykey"`
	Status      int32
	Name        string
	CreatedTime int64
}

type QuizParticipantTab struct {
	ID          int64 `gorm:"primarykey"`
	QuizID      int64
	UserID      int64
	Score       int32
	CreatedTime int64
	UpdatedTime int64
}
