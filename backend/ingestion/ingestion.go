package ingestion

import (
	"context"
	"log"
	"strconv"

	"polychain.capital/config"
	"polychain.capital/db"
	"polychain.capital/rpc"
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
			rpcc:       rpc.New(&network),
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
			log.Println("⌛️ Starting", nw.config.Name, "ingestion workflow")
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

func (s *IngestionService) GetStatus(ctx context.Context) (map[string]NetworkStatus, error) {
	status := make(map[string]NetworkStatus)
	for chanId, nw := range s.nw {
		networkStatus, err := nw.getStatus(ctx)

		if err != nil {
			log.Println("Error getting status for", nw.config.Name, "ingestion worker", err)
			continue
		}
		status[strconv.Itoa(chanId)] = networkStatus
	}
	return status, nil
}
