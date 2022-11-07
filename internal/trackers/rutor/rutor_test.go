package rutor

import (
	"testing"

	"git.rms.local/RacoonMediaServer/rms-torrent/internal/types"
)

func TestSearch(t *testing.T) {
	session := SearchSession{}
	session.Setup(types.SessionSettings{
		UserAgent: "RacoonMediaServer",
	})

	torrents, err := session.Search("Матрица")

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

	torrents, err := session.Search("Матрица")

	if len(torrents) == 0 {
		t.Error("Result must be not empty")
	}

	if err != nil {
		t.Errorf("Error must be nil: %+s", err.Error())
	}

	if _, err := session.Download(torrents[0].DownloadLink); err != nil {
		t.Errorf("Download failed: %+s", err.Error())
	}
}
