package loot

type ResourceManager interface {
	Add(int64) (int64, error)
	Refund(int64) (int64, error)
	Remove(int64) (int64, error)
	Balance() int64
	CanAfford(int64) bool
	Revenue() int64
}

type InMemoryResourceManager struct {
	balance int64
	revenue int64
}

func NewInMemoryResourceManager() (*InMemoryResourceManager, error) {
	return &InMemoryResourceManager{}, nil
}

func (g *InMemoryResourceManager) Add(amount int64) (int64, error) {
	g.balance += amount
	if amount > 0 {
		g.revenue += amount
	}
	return g.balance, nil
}

func (g *InMemoryResourceManager) Refund(amount int64) (int64, error) {
	g.balance += amount
	return g.balance, nil
}

func (g *InMemoryResourceManager) Remove(amount int64) (int64, error) {
	g.balance -= amount
	return g.balance, nil
}

func (g *InMemoryResourceManager) Balance() int64 {
	return g.balance
}

func (g *InMemoryResourceManager) CanAfford(amount int64) bool {
	return g.balance >= amount
}

func (g *InMemoryResourceManager) Revenue() int64 { return g.revenue }
