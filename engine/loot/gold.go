package loot

type GoldManager interface {
	Add(int64) (int64, error)
	Refund(int64) (int64, error)
	Remove(int64) (int64, error)
	Balance() int64
	CanAfford(int64) bool
	Revenue() int64
}

type InMemoryGoldManager struct {
	balance int64
	revenue int64
}

func NewInMemoryGoldManager() (*InMemoryGoldManager, error) {
	return &InMemoryGoldManager{}, nil
}

func (g *InMemoryGoldManager) Add(amount int64) (int64, error) {
	g.balance += amount
	if amount > 0 {
		g.revenue += amount
	}
	return g.balance, nil
}

func (g *InMemoryGoldManager) Refund(amount int64) (int64, error) {
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

func (g *InMemoryGoldManager) Revenue() int64 { return g.revenue }
