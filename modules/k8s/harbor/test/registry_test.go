package test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// registry is an image registry reached with basic auth: Harbor through a port-forward.
type registry struct {
	url      string // http://127.0.0.1:<port>
	password string
}

// do sends a request as admin and returns the body, or an error naming the status when it is
// not the expected one.
func (r registry) do(method, path, contentType string, body []byte, expect int) ([]byte, error) {
	answer, _, err := r.send(method, path, contentType, body, expect)
	return answer, err
}

// send is do that also returns the Location header.
func (r registry) send(method, path, contentType string, body []byte, expect int) ([]byte, string, error) {
	req, err := http.NewRequest(method, r.url+path, bytes.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	req.SetBasicAuth("admin", r.password)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	// A manifest is only served in a media type the client accepts.
	req.Header.Set("Accept", "application/vnd.oci.image.manifest.v1+json, application/json")
	client := http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	answer, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode != expect {
		return nil, "", fmt.Errorf("%s %s returned %d, expected %d: %.300s", method, path, resp.StatusCode, expect, answer)
	}
	return answer, resp.Header.Get("Location"), nil
}

func digest(b []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(b)) }

// testImage builds the smallest image a scanner has something to say about: one layer with
// the release file of an Alpine version, so Trivy recognises the OS. Returns the layer, the
// image config and the manifest.
func testImage(marker string) (layer, config, manifest []byte, err error) {
	var tarball bytes.Buffer
	tw := tar.NewWriter(&tarball)
	for name, content := range map[string]string{"etc/alpine-release": "3.20.0\n", "etc/infralib-test": marker + "\n"} {
		if err = tw.WriteHeader(&tar.Header{Name: name, Mode: 0644, Size: int64(len(content))}); err != nil {
			return
		}
		if _, err = tw.Write([]byte(content)); err != nil {
			return
		}
	}
	if err = tw.Close(); err != nil {
		return
	}
	var zipped bytes.Buffer
	zw := gzip.NewWriter(&zipped)
	if _, err = zw.Write(tarball.Bytes()); err != nil {
		return
	}
	if err = zw.Close(); err != nil {
		return
	}
	layer = zipped.Bytes()
	config = []byte(fmt.Sprintf(`{"architecture":"amd64","os":"linux","config":{},"rootfs":{"type":"layers","diff_ids":[%q]}}`, digest(tarball.Bytes())))
	manifest = []byte(fmt.Sprintf(`{"schemaVersion":2,"mediaType":"application/vnd.oci.image.manifest.v1+json",`+
		`"config":{"mediaType":"application/vnd.oci.image.config.v1+json","digest":%q,"size":%d},`+
		`"layers":[{"mediaType":"application/vnd.oci.image.layer.v1.tar+gzip","digest":%q,"size":%d}]}`,
		digest(config), len(config), digest(layer), len(layer)))
	return
}

// push uploads the blobs and the manifest of an image to <repository>:<tag>.
func (r registry) push(repository, tag string, layer, config, manifest []byte) error {
	for _, blob := range [][]byte{layer, config} {
		_, location, err := r.send(http.MethodPost, fmt.Sprintf("/v2/%s/blobs/uploads/", repository), "", nil, http.StatusAccepted)
		if err != nil {
			return err
		}
		// The upload URL may name the registry's public host, which the port-forward is not:
		// only its path and query are used.
		upload, err := url.Parse(location)
		if err != nil {
			return err
		}
		query := upload.Query()
		query.Set("digest", digest(blob))
		if _, err := r.do(http.MethodPut, upload.Path+"?"+query.Encode(), "application/octet-stream", blob, http.StatusCreated); err != nil {
			return err
		}
	}
	_, err := r.do(http.MethodPut, fmt.Sprintf("/v2/%s/manifests/%s", repository, tag), "application/vnd.oci.image.manifest.v1+json", manifest, http.StatusCreated)
	return err
}

// pull downloads the manifest of <repository>:<tag> and the given blob, and fails when either
// is not what was pushed.
func (r registry) pull(repository, tag string, manifest, layer []byte) error {
	got, err := r.do(http.MethodGet, fmt.Sprintf("/v2/%s/manifests/%s", repository, tag), "", nil, http.StatusOK)
	if err != nil {
		return err
	}
	if !bytes.Equal(got, manifest) {
		return fmt.Errorf("pulled manifest of %s:%s differs from the pushed one: %.300s", repository, tag, got)
	}
	got, err = r.do(http.MethodGet, fmt.Sprintf("/v2/%s/blobs/%s", repository, digest(layer)), "", nil, http.StatusOK)
	if err != nil {
		return err
	}
	if !bytes.Equal(got, layer) {
		return fmt.Errorf("pulled layer of %s:%s differs from the pushed one", repository, tag)
	}
	return nil
}
