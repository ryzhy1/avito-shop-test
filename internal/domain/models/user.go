package models

import (
	"github.com/google/uuid"
	"time"
)

// ну вообще не сказано что монеты у нас могут быть исключительно целыми, а про дробные ничего не говорится.
// ну я хочу чтобы у меня были кофты по 5.99 монет и буду писать код как хочу
// будет хранить количество монет в копейках и умножать на 100 чтобы пользователю выводить приятный глазу вид

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  []byte    `json:"password" db:"password"`
	Coins     int       `json:"coins" db:"coins"` // храним в копейках если что чтбоы было проще хранить и перегонять в большую валюту
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
