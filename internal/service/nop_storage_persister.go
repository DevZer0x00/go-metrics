package service

import "context"

type NopPersister struct{}

func (persister *NopPersister) Init(ctx context.Context) error {
	return nil
}

func (persister *NopPersister) Flush() error {
	return nil
}

func NewNopPersister() *NopPersister {
	return &NopPersister{}
}
