package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "os"

    "github.com/lucas-clemente/quic-go/http3"

    "github.com/notpeko/ytarchive-raw-go/download"
    "github.com/notpeko/ytarchive-raw-go/log"
)

func printResult(logger *log.Logger, res *download.DownloadResult) {
    if len(res.LostSegments) > 0 {
        logger.Warnf("Lost %d segment(s) %v out of %d", len(res.LostSegments), res.LostSegments, res.TotalSegments)
    }
    if res.Error != nil {
        logger.Errorf("Download task failed: %v", res.Error)
    } else {
        logger.Info("Download succeeded")
    }
}

func createMergedFile(which string) string {
    f, err := ioutil.TempFile(tempDir, fmt.Sprintf("merged-%s.%s.", fregData.Metadata.Id, which))
    if err != nil {
        log.Fatalf("Unable to create merged file for %s: %v", which, err)
    }
    f.Close()
    return f.Name()
}

func main() {
    if tempDir == "" {
        var err error
        tempDir, err = ioutil.TempDir("", fmt.Sprintf("ytarchive-%s-", fregData.Metadata.Id))
        if err != nil {
            log.Fatalf("Unable to create temp dir: %v", err)
        }
        log.Info("Storing temporary files in %s", tempDir)
    } else {
        if err := os.MkdirAll(tempDir, 0755); err != nil {
            log.Fatalf("Unable to create temp dir at '%s': %v", tempDir, err)
        }
    }

    client := &http.Client {
        Transport: &http3.RoundTripper{},
    }

    audioTask := &download.DownloadTask {
        Client:         client,
        DeleteSegments: !keepFiles,
        FailThreshold:  failThreshold,
        Logger:         log.New("audio.0"),
        MergeFile:      createMergedFile("audio"),
        QueueMode:      queueMode,
        RetryThreshold: retryThreshold,
        SegmentDir:     tempDir,
        Threads:        threads,
        Url:            fregData.BestAudio(),
    }
    videoTask := &download.DownloadTask {
        Client:         client,
        DeleteSegments: !keepFiles,
        FailThreshold:  failThreshold,
        Logger:         log.New("video.0"),
        MergeFile:      createMergedFile("video"),
        QueueMode:      queueMode,
        RetryThreshold: retryThreshold,
        SegmentDir:     tempDir,
        Threads:        threads,
        Url:            fregData.BestVideo(),
    }

    audioTask.Start()
    videoTask.Start()

    audioRes := audioTask.Wait()
    videoRes := videoTask.Wait()

    printResult(audioTask.Logger, audioRes)
    printResult(videoTask.Logger, videoRes)
}

