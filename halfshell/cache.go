package halfshell

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"strings"
	"time"
)

type Cache struct {
	Enabled bool
	Folder  string
	MaxSize uint64
	Logger  *Logger
}

type CacheFile struct {
	FullName  string
	Extention string
	CacheName string
}

func NewCacheWithConfig(config *ServerConfig) *Cache {
	return &Cache{
		Enabled: config.CacheEnabled,
		Folder:  config.CacheFolder,
		MaxSize: config.CacheMaxSize,
		Logger:  NewLogger("cache"),
	}
}

func (c *Cache) getImage(imagePath string, dimensions ImageDimensions) *Image {
	var image *Image
	f, err := c.getFileHash(imagePath, dimensions)

	if err == nil {
		if info, err := os.Stat(f.CacheName); !os.IsNotExist(err) && (time.Now().Local().Sub(info.ModTime()).Seconds() < 604800) {
			b, err := ioutil.ReadFile(f.CacheName)
			if err == nil || len(b) > 0 {
				image = &Image{
					Bytes:    b,
					MimeType: mime.TypeByExtension(f.Extention),
				}
			}
		}
	}
	return image
}

func (c *Cache) write(imagePath string, dimensions ImageDimensions, image []byte) {
	f, err := c.getFileHash(imagePath, dimensions)
	if err == nil {
		err := ioutil.WriteFile(f.CacheName, image, 0644)
		if err != nil {
			c.Logger.Debug("Error while writing the file %v", f.CacheName)
		}
	}
}

func (c *Cache) getFileHash(path string, dimensions ImageDimensions) (*CacheFile, error) {
	var cacheFile CacheFile
	var err error
	fileNameAray := strings.Split(path, ".")
	if len(fileNameAray) > 1 {
		hasher := md5.New()
		hasher.Write([]byte(fmt.Sprintf("%v_%v_%v", fileNameAray[len(fileNameAray)-2], dimensions.Width, dimensions.Height)))
		hash := fmt.Sprintf("%v.%v", hex.EncodeToString(hasher.Sum(nil)), fileNameAray[len(fileNameAray)-1])
		extention := fileNameAray[len(fileNameAray)-1]
		cacheFile = CacheFile{FullName: hash, Extention: extention, CacheName: fmt.Sprintf("%v/%v", c.Folder, hash)}
	} else {
		err = errors.New("Error while creating the hash")
	}
	return &cacheFile, err
}
