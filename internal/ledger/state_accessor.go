package ledger

import (
	"crypto/sha256"
	"encoding/json"
	"sort"

	"github.com/meshplus/bitxhub-kit/types"
	"github.com/meshplus/bitxhub-model/pb"
)

var _ Ledger = (*ChainLedger)(nil)

// GetOrCreateAccount get the account, if not exist, create a new account
func (l *ChainLedger) GetOrCreateAccount(addr types.Address) *Account {
	h := addr.Hex()
	value, ok := l.accounts[h]
	if ok {
		return value
	}

	account := l.GetAccount(addr)
	l.accounts[addr.Hex()] = account

	return account
}

// GetAccount get account info using account Address, if not found, create a new account
func (l *ChainLedger) GetAccount(addr types.Address) *Account {
	account := newAccount(l.ldb, addr)

	if data := l.ldb.Get(compositeKey(accountKey, addr.Hex())); data != nil {
		account.originAccount = &innerAccount{}
		if err := account.originAccount.Unmarshal(data); err != nil {
			panic(err)
		}
	}

	return account
}

// GetBalanec get account balance using account Address
func (l *ChainLedger) GetBalance(addr types.Address) uint64 {
	account := l.GetOrCreateAccount(addr)
	return account.GetBalance()
}

// SetBalance set account balance
func (l *ChainLedger) SetBalance(addr types.Address, value uint64) {
	account := l.GetOrCreateAccount(addr)
	account.SetBalance(value)
}

// GetState get account state value using account Address and key
func (l *ChainLedger) GetState(addr types.Address, key []byte) (bool, []byte) {
	account := l.GetOrCreateAccount(addr)
	return account.GetState(key)
}

// SetState set account state value using account Address and key
func (l *ChainLedger) SetState(addr types.Address, key []byte, v []byte) {
	account := l.GetOrCreateAccount(addr)
	account.SetState(key, v)
}

// SetCode set contract code
func (l *ChainLedger) SetCode(addr types.Address, code []byte) {
	account := l.GetOrCreateAccount(addr)
	account.SetCodeAndHash(code)
}

// GetCode get contract code
func (l *ChainLedger) GetCode(addr types.Address) []byte {
	account := l.GetOrCreateAccount(addr)
	return account.Code()
}

// GetNonce get account nonce
func (l *ChainLedger) GetNonce(addr types.Address) uint64 {
	account := l.GetOrCreateAccount(addr)
	return account.GetNonce()
}

// SetNonce set account nonce
func (l *ChainLedger) SetNonce(addr types.Address, nonce uint64) {
	account := l.GetOrCreateAccount(addr)
	account.SetNonce(nonce)
}

// QueryByPrefix query value using key
func (l *ChainLedger) QueryByPrefix(addr types.Address, prefix string) (bool, [][]byte) {
	account := l.GetOrCreateAccount(addr)
	return account.Query(prefix)
}

func (l *ChainLedger) Clear() {
	l.events = make(map[string][]*pb.Event, 10)
	l.accounts = make(map[string]*Account)
}

// Commit commit the state
func (l *ChainLedger) Commit(height uint64) (types.Hash, error) {
	var dirtyAccountData []byte
	var journals []*journal
	sortedAddr := make([]string, 0, len(l.accounts))
	accountData := make(map[string][]byte)
	ldbBatch := l.ldb.NewBatch()

	for addr, account := range l.accounts {
		journal := account.getJournalIfModified(ldbBatch)
		if journal != nil {
			sortedAddr = append(sortedAddr, addr)
			accountData[addr] = account.getDirtyData()
			journals = append(journals, journal)
		}
	}

	sort.Strings(sortedAddr)
	for _, addr := range sortedAddr {
		dirtyAccountData = append(dirtyAccountData, accountData[addr]...)
	}
	dirtyAccountData = append(dirtyAccountData, l.prevJournalHash[:]...)
	journalHash := sha256.Sum256(dirtyAccountData)

	blockJournal := BlockJournal{
		Journals:    journals,
		ChangedHash: journalHash,
	}

	data, err := json.Marshal(blockJournal)
	if err != nil {
		return [32]byte{}, err
	}

	ldbBatch.Put(compositeKey(journalKey, height), data)

	ldbBatch.Commit()

	l.height = height
	l.prevJournalHash = journalHash
	l.Clear()

	return journalHash, nil
}

// Version returns the current version
func (l *ChainLedger) Version() uint64 {
	return l.height
}
