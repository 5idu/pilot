package file

import (
	"log"
	"os"
	"path/filepath"

	"github.com/5idu/pilot/pkg/util/xfile"
	"github.com/5idu/pilot/pkg/util/xgo"
	"github.com/5idu/pilot/pkg/xlog"

	"github.com/fsnotify/fsnotify"
)

// fileDataSource file provider.
type fileDataSource struct {
	path        string
	dir         string
	enableWatch bool
	changed     chan struct{}
}

// NewDataSource returns new fileDataSource.
func NewDataSource(path string, watch bool) *fileDataSource {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		xlog.Panic("new datasource", xlog.Any("err", err))
	}

	dir := xfile.CheckAndGetParentDir(absolutePath)
	ds := &fileDataSource{path: absolutePath, dir: dir, enableWatch: watch}
	if watch {
		ds.changed = make(chan struct{}, 1)
		xgo.Go(ds.watch)
	}
	return ds
}

// ReadConfig ...
func (fp *fileDataSource) ReadConfig() (content []byte, err error) {
	return os.ReadFile(fp.path)
}

// Close ...
func (fp *fileDataSource) Close() error {
	close(fp.changed)
	return nil
}

// IsConfigChanged ...
func (fp *fileDataSource) IsConfigChanged() <-chan struct{} {
	return fp.changed
}

// Watch file and automate update.
func (fp *fileDataSource) watch() {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		xlog.Fatal("new file watcher", xlog.String("mod", "file datasource"), xlog.Any("err", err))
	}

	defer w.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-w.Events:
				xlog.Debug("read watch event",
					xlog.String("mod", "file datasource"),
					xlog.String("event", filepath.Clean(event.Name)),
					xlog.String("path", filepath.Clean(fp.path)),
				)
				// we only care about the config file with the following cases:
				// 1 - if the config file was modified or created
				// 2 - if the real path to the config file changed
				const writeOrCreateMask = fsnotify.Write | fsnotify.Create
				if event.Op&writeOrCreateMask != 0 && filepath.Clean(event.Name) == filepath.Clean(fp.path) {
					log.Println("modified file: ", event.Name)
					select {
					case fp.changed <- struct{}{}:
					default:
					}
				}
			case err := <-w.Errors:
				// log.Println("error: ", err)
				xlog.Error("read watch error", xlog.String("mod", "file datasource"), xlog.Any("err", err))
			}
		}
	}()

	err = w.Add(fp.dir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
