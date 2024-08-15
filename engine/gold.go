package engine

type GoldManager interface {
	Add(int64) (int64, error)
	Remove(int64) (int64, error)
	Balance() int64
	CanAfford(int64) bool
}

type InMemoryGoldManager struct {
	balance int64
}

func NewInMemoryGoldManager() (*InMemoryGoldManager, error) {
	return &InMemoryGoldManager{}, nil
}

func (g *InMemoryGoldManager) Add(amount int64) (int64, error) {
	g.balance += amount
	return g.balance, nil
}

func (g *InMemoryGoldManager) Remove(amount int64) (int64, error) {
	g.balance -= amount
	return g.balance, nil
}

func (g *InMemoryGoldManager) Balance() int64 {
	return g.balance
}

func (g *InMemoryGoldManager) CanAfford(amount int64) bool {
	return g.balance >= amount
}
