package tx

type TxMethod int

const (
	SELECT      TxMethod = 10
	SELECT_OK   TxMethod = 11
	COMMIT      TxMethod = 20
	COMMIT_OK   TxMethod = 21
	ROLLBACK    TxMethod = 30
	ROLLBACK_OK TxMethod = 31
)
