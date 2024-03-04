package florist

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"
)

type Hash int

const (
	SHA256 Hash = iota + 1
)

// NetFetch uses client to download url to dstDir, returning the path of the downloaded
// file.
// Directory dstDir must exist.
// If after the download the hash doesn't match, it will return an error.
// If the file in dstDir exists and the hash matches, it will not be
// redownloaded.
func NetFetch(client *http.Client, url string, hashType Hash, hash string, dstDir string) (string, error) {
	log := Log().With("url", url)

	if len(url) == 0 {
		return "", fmt.Errorf("NetFetch: empty url")
	}
	hasher := sha256.New()

	// If file exists and the hash matches, just return.
	dstPath := path.Join(dstDir, path.Base(url))
	fi, err := os.Open(dstPath)
	if err == nil {
		if _, err := io.Copy(hasher, fi); err != nil {
			return "", fmt.Errorf("NetFetch: reading %s: %w", dstPath, err)
		}
		have := hex.EncodeToString(hasher.Sum(nil))
		if have == hash {
			log.Debug("file exist locally, hash matches, skipping download",
				"file", dstPath)
			return dstPath, nil
		}
	}

	if err := Mkdir(dstDir, User().Username, 0775); err != nil {
		return "", fmt.Errorf("NetDetch: %w", err)
	}
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("NetDetch: %w", err)
	}
	defer dst.Close()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("NetDetch: %w", err)
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("NetDetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("NetFetch: received %d %s (GET %s)",
			resp.StatusCode, http.StatusText(resp.StatusCode), url)
	}

	hasher.Reset()
	mw := io.MultiWriter(dst, hasher)

	if _, err := io.Copy(mw, resp.Body); err != nil {
		return "", fmt.Errorf("NetFetch: saving %s to %s: %w", url, dstPath, err)
	}
	elapsed := time.Since(start).Round(time.Millisecond)
	log.Debug("downloaded", "elapsed", elapsed)

	have := hex.EncodeToString(hasher.Sum(nil))
	if have != hash {
		return "", fmt.Errorf("NetFetch: hash mismatch: have: %s; want: %s", have, hash)
	}
	log.Debug("", "dstPath", dstPath)

	return dstPath, nil
}
