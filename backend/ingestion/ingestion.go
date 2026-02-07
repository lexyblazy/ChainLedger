package ingestion

import (
	"context"
	"log"

	"polychain.capital/config"
	"polychain.capital/db"
)

type IngestionService struct {
	db           *db.DB
	config       *config.Config
	nw           map[int]*NetworkWorker
	WorkerErrors map[int]chan error
}

func New(db *db.DB, config *config.Config) *IngestionService {
	workers := make(map[int]*NetworkWorker)
	workerErrors := make(map[int]chan error)

	for _, netCfg := range config.Networks {
		network := netCfg
		workers[network.ChainID] = &NetworkWorker{
			db:         db,
			rpcc:       NewRPCClient(&network),
			config:     &network,
			addressSet: make(map[string]bool),
		}

		workerErrors[network.ChainID] = make(chan error, 1)

	}

	return &IngestionService{
		db:           db,
		nw:           workers,
		config:       config,
		WorkerErrors: workerErrors,
	}
}

func (s *IngestionService) Start(ctx context.Context) error {
	log.Println("Ingestion service started")

	for _, nw := range s.nw {
		go func(nw *NetworkWorker) {
			log.Println("⌛️ Starting", nw.config.Name, "ingestion worker")
			err := nw.start(ctx)
			if err != nil {
				s.WorkerErrors[nw.config.ChainID] <- err
			}
		}(nw)
	}

	return nil
}

func (s *IngestionService) Cleanup(ctx context.Context) {
	log.Println("🧹 Cleaning up ingestion workers...")
	for chainID, errCh := range s.WorkerErrors {
		err := <-errCh
		log.Println("❌ Stopped ", s.nw[chainID].config.Name, "ingestion worker. Reason: ", err)
		close(errCh)
	}

}
