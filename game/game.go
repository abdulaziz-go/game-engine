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

func (b *Board) Reset() {
	b.Cells = [9]byte{}
	b.Turn = tlv.X
	b.Winner = tlv.EMPTY
}

func (b *Board) MakeMove(pos int, symbol byte) bool {
	if pos > 8 || b.Cells[pos] != tlv.EMPTY || b.Turn != symbol || b.Winner != tlv.EMPTY {
		return false
	}

	b.Cells[pos] = symbol
	b.Winner = b.checkWinner()

	if b.Winner == tlv.EMPTY {
		if b.Turn == tlv.X {
			b.Turn = tlv.O
		} else {
			b.Turn = tlv.X
		}
	}

	return true
}

func (b *Board) checkWinner() byte {
	// Check rows, columns, diagonals
	lines := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // columns
		{0, 4, 8}, {2, 4, 6}, // diagonals
	}

	for _, line := range lines {
		if b.Cells[line[0]] != tlv.EMPTY &&
			b.Cells[line[0]] == b.Cells[line[1]] &&
			b.Cells[line[0]] == b.Cells[line[2]] {
			return b.Cells[line[0]]
		}
	}

	// Check for draw
	full := true
	for _, cell := range b.Cells {
		if cell == tlv.EMPTY {
			full = false
			break
		}
	}
	if full {
		return 3 // draw
	}

	return tlv.EMPTY
}
