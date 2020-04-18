package neo

// Starter defines methods for starting a transaction
type Starter interface {
	Begin() (Transactioner, error)
}

// Transactioner outlines requirements for holding a transaction
type Transactioner interface {
	Commit() error
	Rollback() error
}
