package server

import (
	"github.com/joeycumines/go-httphelpers/demo/server/models"
	"sync"
	"errors"
)

type (
	MovieStore interface {
		Delete(ID int) error
		Load(ID int) (movie *models.Movie, ok bool, err error)
		LoadOrStore(ID int, movie *models.Movie) (actual *models.Movie, loaded bool, err error)
		Range(fn func(ID int, movie *models.Movie) bool) error
		Store(ID int, movie *models.Movie) error
		Create(movie *models.Movie) (created *models.Movie, err error)
	}

	movieStore struct {
		mutex sync.Mutex
		id    int
		m     sync.Map
	}
)

func NewMemMovieStore() MovieStore {
	return new(movieStore)
}

func (m *movieStore) Delete(ID int) error {
	panic("implement me")
}

func (m *movieStore) Load(ID int) (movie *models.Movie, ok bool, err error) {
	v, ok := m.m.Load(ID)
	if ok {
		return v.(*models.Movie), true, nil
	}
	return nil, false, nil
}

func (m *movieStore) Range(fn func(ID int, movie *models.Movie) bool) error {
	if fn == nil {
		return errors.New("server.movieStore.Range nil fn")
	}
	m.m.Range(func(key, value interface{}) bool {
		return fn(key.(int), value.(*models.Movie))
	})
	return nil
}

func (m *movieStore) Store(ID int, movie *models.Movie) error {
	panic("implement me")
}

func (m *movieStore) LoadOrStore(ID int, movie *models.Movie) (*models.Movie, bool, error) {
	actual, loaded := m.m.LoadOrStore(ID, movie)
	return actual.(*models.Movie), loaded, nil
}

func (m *movieStore) Create(movie *models.Movie) (*models.Movie, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	ID := m.id
	toStore := new(models.Movie)
	*toStore = *movie
	toStore.ID = ID
	if actual, loaded, err := m.LoadOrStore(ID, toStore); err != nil {
		return nil, err
	} else if loaded || actual != toStore {
		return nil, errors.New("failed to store")
	}
	m.id++
	return toStore, nil
}
