package ctl

import (
	"log"
	"tumblr/circuit/use/circuit"
	"tumblr/circuit/use/durablefs"
	"tumblr/circuit/use/n"
)

func Remove(dfile string) error {
	return durablefs.Remove(dfile)
}

func Kill(dfile string) error {
	chk, err := ReadCheckpoint(dfile)
	if err != nil {
		return circuit.NewError("Problem reading checkpoint (%s)", err)
	}
	if err = KillCheckpoint(chk); err != nil {
		return circuit.NewError("Problems killing shards (%s)", err)
	}
	return Remove(dfile)
}

func KillCheckpoint(chk *Checkpoint) error {
	var err error
	for i, worker := range chk.Workers {
		println("s", i)
		if e := n.Kill(worker.Runtime); err != nil {
			log.Printf("Problem killing server %s (%s)", worker.Runtime, e)
			err = e
		}
	}
	return err
}