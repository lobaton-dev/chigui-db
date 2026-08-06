package drivers

import (
	"fmt"
	"sync"

	"github.com/lobaton-dev/chigui-db/internal/models"
)

type Registry struct {
	mu   sync.RWMutex
	data map[models.DBType]Driver
}

func NewRegistry() *Registry {
	return &Registry{
		data: make(map[models.DBType]Driver),
	}
}

func (r *Registry) Register(dt models.DBType, d Driver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[dt] = d
}

func (r *Registry) Get(dt models.DBType) (Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.data[dt]
	if !ok {
		return nil, fmt.Errorf("unsupported database type %s", dt)
	}
	return d, nil
}

func (r *Registry) List() []models.DBType {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]models.DBType, 0, len(r.data))
	for dt := range r.data {
		result = append(result, dt)
	}
	return result
}

func (r *Registry) RegisterAll() {
	r.Register(models.DBPostgreSQL, &PostgresDriver{})
	r.Register(models.DBMySQL, &MySQLDriver{})
	r.Register(models.DBSQLite, &SQLiteDriver{})
	r.Register(models.DBMSSQL, &MSSQLDriver{})
	r.Register(models.DBOracle, &OracleDriver{})
}
