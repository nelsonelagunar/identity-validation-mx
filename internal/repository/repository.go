package repository

type Repository interface {
	FindID(id string) interface{}
}

type BaseRepository struct{}