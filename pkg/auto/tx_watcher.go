package auto

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"time"
)

type TxWatcher struct {
	PendingTxCh chan *types.Transaction
	ethClient   *ethclient.Client
	quitCh      chan bool
	startedAt   time.Time
	counter     uint
}

func NewTxWatcher(ethClient *ethclient.Client) *TxWatcher {
	return &TxWatcher{
		PendingTxCh: make(chan *types.Transaction, 1000),
		ethClient:   ethClient,
		quitCh:      make(chan bool),
	}
}

func (tw *TxWatcher) Start() {
	tw.startedAt = time.Now()
	go func() {
		defer close(tw.PendingTxCh)
		for {
			select {
			case tx := <-tw.PendingTxCh:
				tw.counter += 1
				if 0 == tw.counter%10 {
					logrus.Debugf("TxW: checking on TX %s (%d in channel)", tx.Hash().Hex(), len(tw.PendingTxCh))
					var receipt *types.Receipt = nil
					sleep := time.Millisecond
					start := time.Now()
					for receipt == nil {
						receipt, _ = tw.ethClient.TransactionReceipt(context.Background(), tx.Hash())
						if nil == receipt {
							time.Sleep(sleep)
							sleep *= 2
						} else {
							elapsed := time.Now().Sub(tw.startedAt)
							logrus.Debugf("TxW: TX %s found in block %s after %dms.", tx.Hash().Hex(),
								receipt.BlockNumber.String(), time.Now().Sub(start).Milliseconds())
							logrus.Infof("TxW: %d in %.0f seconds (%.2f/sec, %d pending)",
								tw.counter, elapsed.Seconds(), float64(tw.counter)/elapsed.Seconds(), len(tw.PendingTxCh))
						}
					}
				}
			case <-tw.quitCh:
				logrus.Infof("TxW: quitting with %d in channel", len(tw.PendingTxCh))
				return
			}
		}
	}()
}
