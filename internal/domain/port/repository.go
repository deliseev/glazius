package port

import "github.com/deliseev/glazius/internal/domain/entity"

type SeriesRepository interface {
	Save(series entity.Series) error
	List() ([]entity.Series, error)
	Get(id string) (entity.Series, error)
	Delete(id string) error
}
