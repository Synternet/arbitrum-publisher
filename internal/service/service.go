package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"arbitrum-publisher/pkg/ipc"
	svcn "arbitrum-publisher/pkg/nats"
)

type SubjectConstants struct {
	StreamedHeader     string
	StreamedBlock      string
	StreamedTx         string
	StreamedTxLogEvent string
	StreamedTxMemPool  string
	SteramedTraceCall  string
}

func NewSubjectConstants(prefix string, network string) *SubjectConstants {
	if network == "" {
		return &SubjectConstants{
			StreamedHeader:     fmt.Sprintf("%s.arbitrum.header", prefix),
			StreamedBlock:      fmt.Sprintf("%s.arbitrum.block", prefix),
			StreamedTx:         fmt.Sprintf("%s.arbitrum.tx", prefix),
			StreamedTxLogEvent: fmt.Sprintf("%s.arbitrum.log-event", prefix),
			StreamedTxMemPool:  fmt.Sprintf("%s.arbitrum.mempool", prefix),
			SteramedTraceCall:  fmt.Sprintf("%s.arbitrum.trace_call", prefix),
		}
	}
	return &SubjectConstants{
		StreamedHeader:     fmt.Sprintf("%s.arbitrum.%s.header", prefix, network),
		StreamedBlock:      fmt.Sprintf("%s.arbitrum.%s.block", prefix, network),
		StreamedTx:         fmt.Sprintf("%s.arbitrum.%s.tx", prefix, network),
		StreamedTxLogEvent: fmt.Sprintf("%s.arbitrum.%s.log-event", prefix, network),
		StreamedTxMemPool:  fmt.Sprintf("%s.arbitrum.%s.mempool", prefix, network),
		SteramedTraceCall:  fmt.Sprintf("%s.arbitrum.%s.trace_call", prefix, network),
	}
}

type Service struct {
	ctx      context.Context
	ipc      *ipc.Ipc
	nats     *svcn.NatsService
	network  string
	subjects *SubjectConstants
	headers  chan *types.Header
	logs     chan types.Log
	ErrCh    chan error
}

func New(ctx context.Context, ipc *ipc.Ipc, s *svcn.NatsService, p string, n string) *Service {
	subjects := NewSubjectConstants(p, n)
	return &Service{
		ctx:      ctx,
		ipc:      ipc,
		nats:     s,
		network:  n,
		subjects: subjects,
		headers:  make(chan *types.Header),
		logs:     make(chan types.Log),
		ErrCh:    make(chan error),
	}
}

func (s *Service) run() {

	defer close(s.ErrCh)
	go s.subscribeNewHeaders()
	go s.subscribeNewLog()
	go s.ipc.MonitorPendingTransactions()

	for {
		select {
		case <-s.ctx.Done():
			return
		case txPoolMessages := <-s.ipc.TxMessagesCh:
			err := s.nats.Publish(s.ctx, s.subjects.StreamedTxMemPool, txPoolMessages.AsJSON())
			if err != nil {
				log.Println(err)
			}
		case txTransactionBlock := <-s.ipc.TxMessageBlock:
			err := s.nats.Publish(s.ctx, s.subjects.StreamedTx, txTransactionBlock.AsJSON())
			if err != nil {
				log.Println(err)
			}
		case txTraceCall := <-s.ipc.TxTraceCalls:
			err := s.nats.Publish(s.ctx, s.subjects.SteramedTraceCall, txTraceCall.AsJSON())
			if err != nil {
				log.Println(err)
			}
		case ipcError := <-s.ipc.IpcErrorCh:
			log.Println(ipcError)
		}
	}
}

func (s *Service) subscribeNewHeaders() {
	subHeaders, err := s.ipc.EthClient.SubscribeNewHead(s.ctx, s.headers)
	if err != nil {
		s.ErrCh <- err

		return
	}
	log.Println("subscribed to newHeader")
	for {
		select {
		case errHeaders := <-subHeaders.Err():
			s.ErrCh <- errHeaders
		case header := <-s.headers:
			rhead, rblock, err := s.ipc.ProcessAndPrepareBlock(header)
			if err != nil {
				log.Printf("Processing header data: %s", err.Error())

				continue
			}

			log.Printf("Captured Head: %s\r\n", header.Hash().Hex())

			err = s.nats.PublishAsJSON(s.ctx, s.subjects.StreamedHeader, rhead)
			if err != nil {
				log.Println(err)
			}

			err = s.nats.PublishAsJSON(s.ctx, s.subjects.StreamedBlock, rblock)
			if err != nil {
				log.Println(err)
			}
		}
	}

}

func (s *Service) subscribeNewLog() {
	filter := ethereum.FilterQuery{
		Addresses: []common.Address{}, // filter all addresses
	}
	subTxPool, err := s.ipc.EthClient.SubscribeFilterLogs(s.ctx, filter, s.logs)
	if err != nil {
		s.ErrCh <- err
		return
	}
	log.Println("subscribed to newLog")
	for {
		select {
		case errTxPool := <-subTxPool.Err():
			s.ErrCh <- errTxPool
		case txLog := <-s.logs:
			if txLog.Removed { // Transaction is no longer in tx-pool log
				continue
			}
			errLog := s.nats.PublishAsJSON(s.ctx, s.subjects.StreamedTxLogEvent, txLog)
			if errLog != nil {
				s.ErrCh <- errLog
				continue
			}
			fmt.Println("log message for transaction:", txLog.TxHash.Hex())
		}
	}
}

func (s *Service) Run() <-chan error {
	go s.run()

	return s.ErrCh
}
