package txdb

type RunInTx func() error

func NoopInTx() error {
	return nil
}
