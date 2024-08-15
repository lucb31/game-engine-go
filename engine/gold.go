package engine

import "fmt"

type GoldManager interface {
	Add(int64) (int64, error)
	Remove(int64) (int64, error)
	Balance() (int64, error)
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
	fmt.Printf("Adding %d gold. New balance is %d \n", amount, g.balance)
	return g.balance, nil
}

func (g *InMemoryGoldManager) Remove(amount int64) (int64, error) {
	g.balance -= amount
	fmt.Printf("Removing %d gold. New balance is %d \n", amount, g.balance)
	return g.balance, nil
}

func (g *InMemoryGoldManager) Balance() (int64, error) {
	return g.balance, nil
}

func (g *InMemoryGoldManager) CanAfford(amount int64) bool {
	return g.balance >= amount
}
