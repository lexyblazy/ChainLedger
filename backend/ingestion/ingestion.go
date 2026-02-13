package ingestion

import (
	"context"
	"log"
	"strconv"
	"sync"

	"chainledger/config"
	"chainledger/db"
)

type IngestionService struct {
	db           *db.DB
	config       *config.Config
	nw           map[int]*NetworkWorker
	WorkerErrors map[int]chan error
}

type StatusReader struct {
	db        *db.DB
	config    *config.Config
	nwReaders map[int]*NetworkReader
}

// exposes a read only layer of the ingestion service
func NewStatusReader(db *db.DB, config *config.Config) *StatusReader {

	nwReaders := make(map[int]*NetworkReader)

	for _, netCfg := range config.Networks {
		network := netCfg
		nwReader := NewNetworkReader(db, &network)
		nwReaders[network.ChainID] = nwReader
	}

	return &StatusReader{
		db:        db,
		config:    config,
		nwReaders: nwReaders,
	}
}

// exposes the full ingestion service (worker + reader)
func New(db *db.DB, config *config.Config) *IngestionService {
	workers := make(map[int]*NetworkWorker)
	workerErrors := make(map[int]chan error)

	for _, netCfg := range config.Networks {
		network := netCfg
		nr := NewNetworkReader(db, &network)
		workers[network.ChainID] = &NetworkWorker{
			db:         db,
			config:     &network,
			addressSet: make(map[string]bool),
			nr:         nr,
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
			err := nw.start(ctx, s.config.SyncBlocks)
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

func (s *StatusReader) GetStatus(ctx context.Context) (map[string]NetworkStatus, error) {
	status := make(map[string]NetworkStatus)

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	for _, nwReader := range s.nwReaders {
		wg.Add(1)
		go func(nr *NetworkReader) {
			defer wg.Done()
			networkStatus, err := nr.getStatus(ctx)
			if err != nil {
				log.Println("Error getting status for", nr.config.Name, "ingestion worker", err)
				return
			}
			mu.Lock()
			status[strconv.Itoa(nr.config.ChainID)] = networkStatus
			mu.Unlock()
		}(nwReader)
	}

	wg.Wait()
	return status, nil
}
