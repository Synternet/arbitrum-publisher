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

func NewSubjectConstants(prefix string, publisher string, network string) *SubjectConstants {
	if network == "" {
		return &SubjectConstants{
			StreamedHeader:     fmt.Sprintf("%s.%s.header", prefix, publisher),
			StreamedBlock:      fmt.Sprintf("%s.%s.block", prefix, publisher),
			StreamedTx:         fmt.Sprintf("%s.%s.tx", prefix, publisher),
			StreamedTxLogEvent: fmt.Sprintf("%s.%s.log-event", prefix, publisher),
			StreamedTxMemPool:  fmt.Sprintf("%s.%s.mempool", prefix, publisher),
			SteramedTraceCall:  fmt.Sprintf("%s.%s.trace_call", prefix, publisher),
		}
	}
	return &SubjectConstants{
		StreamedHeader:     fmt.Sprintf("%s.%s.%s.header", prefix, publisher, network),
		StreamedBlock:      fmt.Sprintf("%s.%s.%s.block", prefix, publisher, network),
		StreamedTx:         fmt.Sprintf("%s.%s.%s.tx", prefix, publisher, network),
		StreamedTxLogEvent: fmt.Sprintf("%s.%s.%s.log-event", prefix, publisher, network),
		StreamedTxMemPool:  fmt.Sprintf("%s.%s.%s.mempool", prefix, publisher, network),
		SteramedTraceCall:  fmt.Sprintf("%s.%s.%s.trace_call", prefix, publisher, network),
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

func New(ctx context.Context, ipc *ipc.Ipc, s *svcn.NatsService, p string, publisher string, n string) *Service {
	subjects := NewSubjectConstants(p, publisher, n)
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
	log.Println("Subscribing to newHeader...")
	subHeaders, err := s.ipc.EthClient.SubscribeNewHead(s.ctx, s.headers)
	if err != nil {
		s.ErrCh <- fmt.Errorf("failed to subscribe to newHeader: %w", err)
		return
	}
	log.Println("Subscribed to newHeader")
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
	log.Println("Subscribing to newLog...")
	filter := ethereum.FilterQuery{
		Addresses: []common.Address{}, // filter all addresses
	}
	subTxPool, err := s.ipc.EthClient.SubscribeFilterLogs(s.ctx, filter, s.logs)
	if err != nil {
		s.ErrCh <- fmt.Errorf("failed to subscribe to newLog: %w", err)
		return
	}
	log.Println("Subscribed to newLog")
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
