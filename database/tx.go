package database

type RunInTx func() error

func NopInTx() error {
	return nil
}
