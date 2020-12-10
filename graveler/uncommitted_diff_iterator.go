package graveler

import (
	"context"
	"errors"

	"github.com/treeverse/lakefs/logging"
)

type uncommittedDiffIterator struct {
	committedManager CommittedManager
	list             ValueIterator
	sn               StorageNamespace
	treeID           TreeID
	value            *Diff
	err              error
}

func NewUncommittedDiffIterator(manager CommittedManager, list ValueIterator, sn StorageNamespace, treeItreeID TreeID) DiffIterator {
	return &uncommittedDiffIterator{
		committedManager: manager,
		list:             list,
		sn:               sn,
		treeID:           treeItreeID,
	}
}

func valueExistsInCommitted(ctx context.Context, committedManager CommittedManager, sn StorageNamespace, treeID TreeID, key Key) (bool, error) {
	_, err := committedManager.Get(ctx, sn, treeID, key)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func getDiffType(ctx context.Context, committedManager CommittedManager, sn StorageNamespace, treeID TreeID, key Key, tombstone bool) (DiffType, error) {
	existsInCommitted, err := valueExistsInCommitted(ctx, committedManager, sn, treeID, key)
	if err != nil {
		return 0, err
	}
	diffType := DiffTypeAdded
	if tombstone {
		if !existsInCommitted {
			logging.Default().
				WithFields(logging.Fields{"tree_id": treeID, "storage_namespace": sn, "key": key}).
				Warn("tombstone for a file that does not exist")
		}
		return DiffTypeRemoved
	}
	if existsInCommitted {
		return DiffTypeChanged
	}
	return DiffTypeAdded
}

func (d *uncommittedDiffIterator) Next() bool {
	if !d.list.Next() {
		d.value = nil
		return false
	}
	val := d.list.Value()
	diffType, err := getDiffType(context.Background(), d.committedManager, d.sn, d.treeID, val.Key, val.Value == nil)
	if err != nil {
		d.value = nil
		d.err = err
		return false
	}
	d.value = &Diff{
		Type:  diffType,
		Key:   val.Key,
		Value: val.Value,
	}
	return true
}

func (d *uncommittedDiffIterator) SeekGE(id Key) {
	d.value = nil
	d.list.SeekGE(id)
}

func (d *uncommittedDiffIterator) Value() *Diff {
	return d.value
}

func (d *uncommittedDiffIterator) Err() error {
	return d.err
}

func (d *uncommittedDiffIterator) Close() {
	d.list.Close()
}
