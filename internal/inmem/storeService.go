package inmem

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"
)

type StoreService struct {
	storage  *MemStorage
	filePath string
}

func NewService(storage *MemStorage, filePath string) StoreService {
	return StoreService{
		storage:  storage,
		filePath: filePath,
	}
}

func (s StoreService) Run(interval int) {
	for {
		time.Sleep(time.Second * time.Duration(interval))
		s.Dump()
	}
}

func (s StoreService) Restore(restore bool) error {
	if restore {
		file, err := os.OpenFile(s.filePath, os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		buf := bufio.NewScanner(file)
		items := []metrics.Metrics{}

		for {
			if !buf.Scan() {
				break
			}
			fmt.Println(string(buf.Bytes()))
			data := buf.Bytes()
			item := metrics.Metrics{}
			err := json.Unmarshal(data, &item)
			if err != nil {
				logger.Get().Info(err.Error())
				continue
			}
			items = append(items, item)
		}

		s.storage.SetMetrics(items)
		return nil
	}
	return nil
}

func (s StoreService) Dump() error {
	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	items := s.storage.GetMetrics()
	for _, item := range items {
		row, err := json.Marshal(item)
		if err != nil {
			return err
		}

		file.Write(row)
		file.WriteString("\n")
	}

	return nil
}
