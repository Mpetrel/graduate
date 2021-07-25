package token

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/metadata"
	"strconv"
)

func ExtractUid(ctx context.Context) (uint64, error) {
	meta, ok := metadata.FromServerContext(ctx)
	if !ok {
		return 0, errors.New(4001,"no metadata found", "no metadata found")
	}
	intNum, err := strconv.Atoi(meta.Get("uid"))
	if err != nil {
		return 0, errors.New(4002, "invalid uid", "invalid uid")
	}
	return uint64(intNum), nil
}
