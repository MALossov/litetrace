package ftrace

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func (e *Engine) ReadTracePipe(ctx context.Context) <-chan string {
	channel := make(chan string, 100)

	go func() {
		defer close(channel)

		filePath := filepath.Join(e.tracefsPath, "trace_pipe")
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open trace_pipe: %v\n", err)
			return
		}
		defer file.Close()

		reader := bufio.NewReader(file)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				line, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				select {
				case channel <- line:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return channel
}
