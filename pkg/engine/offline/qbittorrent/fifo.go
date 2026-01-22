package qbittorrent

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	"github.com/RacoonMediaServer/rms-torrent/v4/pkg/engine"
	"go-micro.dev/v4/logger"
	"golang.org/x/sys/unix"
)

const fifoPollInterval = 100 * time.Millisecond

type fifoWatcher struct {
	f      *os.File
	action engine.CompleteAction
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func (w *fifoWatcher) startFifoWatcher(fifoPath string, action engine.CompleteAction) error {
	_, err := os.Stat(fifoPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = unix.Mkfifo(fifoPath, 0666); err != nil {
				return fmt.Errorf("create fifo failed: %w", err)
			}
		} else {
			return err
		}
	}
	f, err := os.OpenFile(fifoPath, os.O_RDONLY|syscall.O_NONBLOCK, 0666)
	if err != nil {
		return fmt.Errorf("open fifo failed: %w", err)
	}

	w.f = f
	w.action = action

	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.watch()
	}()

	return nil
}

func (w *fifoWatcher) watch() {
	for {
		scanner := bufio.NewScanner(w.f)
		for scanner.Scan() {
			if err := w.handleNotification(scanner.Text()); err != nil {
				logger.Error("Handle notification failed: %s", err)
			}
		}
		select {
		case <-w.ctx.Done():
			return
		case <-time.After(fifoPollInterval):
		}
	}
}

func (w *fifoWatcher) stop() error {
	w.cancel()
	w.wg.Wait()
	if w.f != nil {
		if err := w.f.Close(); err != nil {
			return fmt.Errorf("close fifo failed: %w", err)
		}
	}
	return nil
}

const notificationPattern = `%L|%G|%Z|%I|%N`

func (w *fifoWatcher) handleNotification(text string) error {
	logger.Infof("Got notification: %s", text)
	items := strings.Split(text, "|")
	if len(items) < 5 {
		return errors.New("invalid notification format")
	}

	event := events.Notification{
		Sender:    "rms-torrent",
		Kind:      events.Notification_DownloadComplete,
		TorrentID: &items[3],
		ItemTitle: &items[4],
	}
	event.MediaID = &items[1]

	size, err := strconv.ParseUint(items[2], 10, 64)
	if err == nil {
		sz := uint32(float32(size) / 1024. * 1034.)
		event.SizeMB = &sz
	}

	if err := w.action(&event); err != nil {
		return fmt.Errorf("complete action failed: %w", err)
	}
	return nil
}
