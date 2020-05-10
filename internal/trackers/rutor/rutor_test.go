package rutor

import (
	"os"
	"racoondev.tk/gitea/racoon/rms-torrent/internal/types"
	"testing"
)


func TestSearch(t *testing.T) {
	session := SearchSession{}
	session.Setup(types.SessionSettings{
		UserAgent: "RacoonMediaServer",
	})

	torrents, err := session.Search("Матрица", 10)

	if torrents == nil {
		t.Error("Result must be not nil")
	}

	if err != nil {
		t.Errorf("Error must be nil: %+s", err.Error())
	}

	t.Logf("Torrents: %+v", torrents)
}

func TestDownload(t *testing.T) {
	session := SearchSession{}
	session.Setup(types.SessionSettings{
		UserAgent: "RacoonMediaServer",
	})

	torrents, err := session.Search("Матрица", 10)

	if torrents == nil || len(torrents) == 0 {
		t.Error("Result must be not empty")
	}

	if err != nil {
		t.Errorf("Error must be nil: %+s", err.Error())
	}

	if err := session.Download(torrents[0].DownloadLink, "test.torrent"); err != nil {
		t.Errorf("Download failed: %+s", err.Error())
	}

	os.Remove("test.torrent")
}
