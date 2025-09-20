package game

import "gamehomework/tlv"

type Board struct {
	Cells  [9]byte
	Turn   byte
	Winner byte
}

func NewBoard() *Board {
	return &Board{
		Turn:   tlv.X,
		Winner: tlv.EMPTY,
	}
}

func (b *Board) ToBytes() []byte {
	data := make([]byte, 11)
	copy(data[:9], b.Cells[:])
	data[9] = b.Turn
	data[10] = b.Winner
	return data
}
