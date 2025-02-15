package claimtxman

import (
	"context"
	"fmt"
	"time"

	ctmtypes "github.com/fiwallets/zkevm-bridge-service/claimtxman/types"
	"github.com/fiwallets/zkevm-bridge-service/log"
	"github.com/jackc/pgx/v4"
)

type StoreChanges struct {
	AddGroups    []ctmtypes.MonitoredTxGroupDBEntry
	UpdateGroups []ctmtypes.MonitoredTxGroupDBEntry
	UpdateTxs    []ctmtypes.MonitoredTx
}

func NewStoreChanges() *StoreChanges {
	return &StoreChanges{}
}
func (sc *StoreChanges) AddGroup(group ctmtypes.MonitoredTxGroupDBEntry) {
	sc.AddGroups = append(sc.AddGroups, group)
}
func (sc *StoreChanges) UpdateGroup(group ctmtypes.MonitoredTxGroupDBEntry) {
	sc.UpdateGroups = append(sc.UpdateGroups, group)
}
func (sc *StoreChanges) UpdateTx(tx ctmtypes.MonitoredTx) {
	sc.UpdateTxs = append(sc.UpdateTxs, tx)
}

func (sc *StoreChanges) Execute(ctx context.Context, storage StorageCompressedInterface, dbTx pgx.Tx) error {
	for i := range sc.AddGroups {
		log.Debugf("Adding group %d ", sc.AddGroups[i].GroupID)
		err := storage.AddMonitoredTxsGroup(ctx, &sc.AddGroups[i], dbTx)
		if err != nil {
			return fmt.Errorf("storeChanges.Execute error adding MonitoresTxGroup. Err: %w", err)
		}
		log.Infof("Added group %d", sc.AddGroups[i].GroupID)
	}

	for i := range sc.UpdateGroups {
		sc.UpdateGroups[i].UpdatedAt = time.Now()
		err := storage.UpdateMonitoredTxsGroup(ctx, &sc.UpdateGroups[i], dbTx)
		if err != nil {
			return err
		}
		log.Infof("Updated group %d", sc.UpdateGroups[i].GroupID)
	}

	for i := range sc.UpdateTxs {
		if sc.UpdateTxs[i].GroupID != nil {
			log.Debugf("Updating tx deposit_id: %d. Group_id:%d", sc.UpdateTxs[i].DepositID, *sc.UpdateTxs[i].GroupID)
		} else {
			log.Warnf("Updating tx deposit_id: %d. Group_id is nil", sc.UpdateTxs[i].DepositID)
		}
		err := storage.UpdateClaimTx(ctx, sc.UpdateTxs[i], dbTx)
		if err != nil {
			return fmt.Errorf("storeChanges.Execute error UpdateClaimTx. Err: %w", err)
		}
	}
	return nil
}
