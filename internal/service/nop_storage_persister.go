package service

type NopPersister struct{}

func (persister *NopPersister) Init() error {
	return nil
}

func (persister *NopPersister) Flush() error {
	return nil
}

func NewNopPersister() *NopPersister {
	return &NopPersister{}
}
