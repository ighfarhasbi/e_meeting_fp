package entity

type Snacks struct {
	ID       int     `json:"id" db:"snacks_id"`
	Name     string  `json:"name" db:"name"`
	Price    float32 `json:"price" db:"price"`
	Category string  `json:"category" db:"category"`
}
