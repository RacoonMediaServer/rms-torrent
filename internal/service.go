package internal

import (
	"context"
	"fmt"
	proto "racoondev.tk/gitea/racoon/rtorrent/proto"
)

type TorrentService struct {

}

func (service *TorrentService) Search(ctx context.Context, in *proto.SearchRequest, out *proto.SearchResponse) error {
	fmt.Println("Search called")
	return nil
}
