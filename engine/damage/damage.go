package damage

type Attacker interface {
	Power() float64
}

type Defender interface {
	Health() float64
	Armor() float64
	Destroy() error
	SetHealth(float64)
}

func CalculateDamage(atk Attacker, def Defender) (float64, error) {
	damage := atk.Power() - def.Armor()
	return damage, nil
}

func ApplyDamage(atk Attacker, def Defender) error {
	damage, err := CalculateDamage(atk, def)
	if err != nil {
		return err
	}
	newHealth := def.Health() - damage
	def.SetHealth(newHealth)
	if newHealth <= 0.0 {
		return def.Destroy()
	}
	return nil
}
