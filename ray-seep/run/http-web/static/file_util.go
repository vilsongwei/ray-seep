package static

import (
	"os"
	"path/filepath"
	"time"
)

type FileInfo interface {
	Name() string
	Size() int64
	Time() time.Time
}
type defFileInfo struct {
	name  string
	size  int64
	cTime time.Time
}

func (d *defFileInfo) Name() string {
	return d.name
}

func (d *defFileInfo) Size() int64 {
	return d.size
}

func (d *defFileInfo) Time() time.Time {
	return d.cTime
}

// FileExists 判断文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func PathEasyWolk(root string) ([]FileInfo, error) {
	return pathWolk(root, false)
}

func PathLoopWolk(root string) ([]FileInfo, error) {
	return pathWolk(root, true)
}

func pathWolk(root string, isLoop bool) ([]FileInfo, error) {
	var list []FileInfo
	if err := filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if fi == nil {
			return filepath.SkipDir
		}
		if err != nil {
			return err
		}
		list = append(list, &defFileInfo{
			name:  fi.Name(),
			size:  fi.Size(),
			cTime: fi.ModTime(),
		})
		if (!isLoop) && fi.IsDir() && (path != root) {
			return filepath.SkipDir
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return list, nil
}

func IsDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}
