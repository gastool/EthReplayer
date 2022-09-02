package replay

import (
	"errors"
	"gopkg.in/urfave/cli.v1"
	"strconv"
)

var replayCommand = cli.Command{
	Action:      replayCmd,
	Name:        "replay",
	Usage:       "replay blockNumber txIndex",
	ArgsUsage:   "<blockNumber> <txIndex>",
	Description: `replay tx`,
}

func replayCmd(ctx *cli.Context) error {
	if len(ctx.Args()) != 2 {
		return errors.New("invalid blockNumber or txIndex")
	}
	block, _ := strconv.Atoi(ctx.Args()[0])
	tx, _ := strconv.Atoi(ctx.Args()[1])
	return Replay(uint64(block), tx)
}
