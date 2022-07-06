package contrib

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"agent/api/v1/model"
	"agent/internal/pkg/global"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// This is an example implementation of the global.Stream interface
// and its sole purpose is to demonstrate how to access the stream
// of data generated by the agent. It will marshal all messages
// consumed to JSON and write them under $HOME/.cache/agent_stream.log.

var outPath = filepath.Join(global.AgentCacheDir, "agent_stream.log")

type logLine struct {
	Hostname string              `json:"hostname"`
	Mf       *model.MetricFamily `json:"mf,omitempty"`
	Ev       *model.Event        `json:"ev,omitempty"`
}

type jsonProcessor struct{}

func (m *jsonProcessor) Process(msgi interface{}) ([]byte, error) {
	msg, ok := msgi.(*model.Message)
	if !ok {
		return nil, fmt.Errorf("type assertion failed: %T", msg)
	}

	var err error
	line := &logLine{Hostname: global.AgentHostname}

	switch msg.Type {
	case model.MessageType_metric:
		line.Mf = &model.MetricFamily{}
		if err := proto.Unmarshal(msg.Body, line.Mf); err != nil {
			return nil, err
		}
	case model.MessageType_event:
		line.Ev = &model.Event{}
		if err := proto.Unmarshal(msg.Body, line.Ev); err != nil {
			return nil, err
		}
	default:
		zap.S().Errorf("unknown msg type: %v", msg.Type)

		return nil, err
	}

	bytes, err := json.Marshal(&line)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

type fileStream struct {
	*jsonProcessor

	ch chan interface{}
	f  *os.File
}

func newFileStream() (global.Stream, error) {
	f, err := os.OpenFile(outPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o755)
	if err != nil {
		return nil, err
	}

	return &fileStream{jsonProcessor: &jsonProcessor{}, f: f}, nil
}

func (m *fileStream) Start(ctx context.Context, wg *sync.WaitGroup, ch chan interface{}) {
	log := zap.S()

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}

				body, err := m.Process(msg)
				if err != nil {
					log.Errorw("file stream handle error", zap.Error(err))

					continue
				}

				if _, err := m.f.Write(append(body, '\n')); err != nil {
					log.Errorw("write error", zap.Error(err))

					continue
				}
			case <-ctx.Done():
				if err := m.f.Close(); err != nil {
					log.Warnw("close error", zap.Error(err))
				}

				return
			}
		}
	}()
}