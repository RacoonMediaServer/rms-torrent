package rutor

import (
	"racoondev.tk/gitea/racoon/rtorrent/internal/types"
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
