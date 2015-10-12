package board

import (
	"log"
)

type Board struct {
	board [][]int
	size  int
}

func New(size int) *Board {
	b := new(Board)
	b.size = size
	b.board = make([][]int, size+2)
	for y := 0; y < size+2; y++ {
		b.board[y] = make([]int, size+2)
		for x := 0; x < size+2; x++ {
			b.board[y][x] = 0
		}
	}
	return b
}

func New19() *Board { return New(19) }

func New13() *Board { return New(13) }

func New9() *Board { return New(9) }

func PutPos(b *Board, posX int, posY int, which int) bool {
	if b.board[posY][posX] != 0 {
		return false
	}
	b.board[posY][posX] = which
	for _, v := range b.board {
		log.Println(v)
	}
	return true
}

// うまく判定されなかった
func lenCheck(b *Board, x int, y int) bool {
	dx := [3]int{0, 1, 1}
	dy := [3]int{1, 0, 1}
	lenFlag := 1

	for i := 0; i < 3; i++ {
		for j := 1; j <= 4; j++ {
			if b.board[y][x] != b.board[y+j*dy[i]][x+j*dx[i]] {
				lenFlag = 0
				break
			}
		}
		if lenFlag == 1 {
			return true
		}
	}
	return false
}

// 上のがうまく判定されなかったので、超自力
func lenCheck2(b *Board, x int, y int) bool {
	/*縦の検索*/
	line := 1
	for i := 1; i <= 4; i++ {
		if x+i > 12 {
			break
		}
		if b.board[y][x+i] == b.board[y][x] {
			line++
		} else {
			break
		}
		if line == 5 {
			return true
		}
	}
	/*横の検索*/
	line = 1
	for i := 1; i <= 4; i++ {
		if x+i > 12 {
			break
		}
		if b.board[y+i][x] == b.board[y][x] {
			line++
		} else {
			break
		}
		if line == 5 {
			return true
		}
	}
	/* 斜めの検索(右下) */
	line = 1
	for i := 1; i <= 4; i++ {
		if x+i > 12 || y+i > 12 {
			break
		}
		if b.board[y+i][x+i] == b.board[y][x] {
			line++
		} else {
			break
		}
		if line == 5 {
			return true
		}
	}
	/* 斜めの検索(右上) */
	line = 1
	for i := 1; i <= 4; i++ {
		if y-i < 0 || x-i < 0 {
			break
		}
		if b.board[y-i][x+i] == b.board[y][x] {
			line++
		} else {
			break
		}
	}
	if line == 5 {
		return true
	}
	return false
}

func GameEnd(b *Board) bool {
	for i := 0; i < b.size; i++ {
		for j := 0; j < b.size; j++ {
			if b.board[i][j] == 0 {
				continue
			}
			if lenCheck2(b, j, i) {
				if b.board[i][j] == 1 {
					log.Println("黒の勝ちです")
				} else {
					log.Println("白の勝ちです")
				}
				return true
			}
		}
	}
	return false
}
