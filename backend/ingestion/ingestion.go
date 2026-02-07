package ingestion

import (
	"context"
	"log"

	"polychain.capital/config"
	"polychain.capital/db"
)

type IngestionService struct {
	db     *db.DB
	config *config.Config
	nw     map[int]*NetworkWorker
}

func New(db *db.DB, config *config.Config) *IngestionService {
	workers := make(map[int]*NetworkWorker)

	for _, netCfg := range config.Networks {
		network := netCfg
		workers[network.ChainID] = &NetworkWorker{
			db:         db,
			rpcc:       NewRPCClient(&network),
			config:     &network,
			addressSet: make(map[string]bool),
		}
	}

	return &IngestionService{db: db, nw: workers, config: config}
}

func (s *IngestionService) Start(ctx context.Context) error {
	log.Println("Ingestion service started")

	for _, nw := range s.nw {
		go func(nw *NetworkWorker) {
			err := nw.start(ctx)
			if err != nil {
				log.Printf("Error starting %s worker: %v", nw.config.Name, err)
			}
		}(nw)
	}

	return nil
}
