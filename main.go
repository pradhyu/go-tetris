package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	boardWidth  = 10
	boardHeight = 20
)

// Tetromino shapes
var tetrominoes = [][]string{
	{ // I
		".....",
		".....",
		"XXXX.",
		".....",
		".....",
	},
	{ // O
		".....",
		".....",
		".XX..",
		".XX..",
		".....",
	},
	{ // T
		".....",
		".....",
		".XXX.",
		"..X..",
		".....",
	},
	{ // L
		".....",
		".....",
		"XXX..",
		"X....",
		".....",
	},
	{ // J
		".....",
		".....",
		"XXX..",
		"..X..",
		".....",
	},
	{ // S
		".....",
		".....",
		".XX..",
		"XX...",
		".....",
	},
	{ // Z
		".....",
		".....",
		"XX...",
		".XX..",
		".....",
	},
}

type Game struct {
	board        [][]bool
	currentPiece [][]bool
	pieceX       int
	pieceY       int
	score        int
}

func NewGame() *Game {
	board := make([][]bool, boardHeight)
	for i := range board {
		board[i] = make([]bool, boardWidth)
	}
	return &Game{
		board: board,
	}
}

func (g *Game) spawnPiece() bool {
	// Select random tetromino
	shape := tetrominoes[rand.Intn(len(tetrominoes))]
	g.currentPiece = make([][]bool, 5)
	for i := range g.currentPiece {
		g.currentPiece[i] = make([]bool, 5)
		for j, char := range shape[i] {
			g.currentPiece[i][j] = char == 'X'
		}
	}
	g.pieceX = boardWidth/2 - 2
	g.pieceY = 0

	// Check if the new piece can be placed
	if g.checkCollision() {
		return false // Game over
	}
	return true
}

func (g *Game) draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// Draw board
	for y := 0; y < boardHeight; y++ {
		for x := 0; x < boardWidth; x++ {
			char := '.'
			if g.board[y][x] {
				char = '#'
			}
			termbox.SetCell(x*2, y, char, termbox.ColorWhite, termbox.ColorDefault)
		}
	}

	// Draw current piece
	if g.currentPiece != nil {
		for y := 0; y < 5; y++ {
			for x := 0; x < 5; x++ {
				if g.currentPiece[y][x] {
					screenX := (g.pieceX + x) * 2
					screenY := g.pieceY + y
					if screenX >= 0 && screenX < boardWidth*2 && screenY >= 0 && screenY < boardHeight {
						termbox.SetCell(screenX, screenY, '#', termbox.ColorCyan, termbox.ColorDefault)
					}
				}
			}
		}
	}

	// Draw score
	scoreStr := fmt.Sprintf("Score: %d", g.score)
	for i, char := range scoreStr {
		termbox.SetCell(boardWidth*2+2+i, 0, char, termbox.ColorWhite, termbox.ColorDefault)
	}

	termbox.Flush()
}

func (g *Game) moveDown() bool {
	g.pieceY++
	if g.checkCollision() {
		g.pieceY--
		g.mergePiece()
		g.clearLines()
		if !g.spawnPiece() {
			return false // Game over
		}
	}
	return true
}

func (g *Game) moveLeft() {
	g.pieceX--
	if g.checkCollision() {
		g.pieceX++
	}
}

func (g *Game) moveRight() {
	g.pieceX++
	if g.checkCollision() {
		g.pieceX--
	}
}

func (g *Game) checkCollision() bool {
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if g.currentPiece[y][x] {
				boardX := g.pieceX + x
				boardY := g.pieceY + y
				if boardX < 0 || boardX >= boardWidth || boardY >= boardHeight {
					return true
				}
				if boardY >= 0 && g.board[boardY][boardX] {
					return true
				}
			}
		}
	}
	return false
}

func (g *Game) mergePiece() {
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if g.currentPiece[y][x] {
				boardY := g.pieceY + y
				boardX := g.pieceX + x
				if boardY >= 0 && boardY < boardHeight && boardX >= 0 && boardX < boardWidth {
					g.board[boardY][boardX] = true
				}
			}
		}
	}
}

func (g *Game) clearLines() {
	for y := boardHeight - 1; y >= 0; y-- {
		full := true
		for x := 0; x < boardWidth; x++ {
			if !g.board[y][x] {
				full = false
				break
			}
		}
		if full {
			g.score += 100
			// Move all lines above down
			for moveY := y; moveY > 0; moveY-- {
				copy(g.board[moveY], g.board[moveY-1])
			}
			// Clear top line
			for x := 0; x < boardWidth; x++ {
				g.board[0][x] = false
			}
			y++ // Check the same line again
		}
	}
}

func (g *Game) rotate() {
	// Create a new 5x5 matrix for the rotated piece
	rotated := make([][]bool, 5)
	for i := range rotated {
		rotated[i] = make([]bool, 5)
	}

	// Rotate 90 degrees clockwise around the center
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			rotated[y][x] = g.currentPiece[4-x][y]
		}
	}

	// Save the current position and piece
	oldPiece := g.currentPiece
	g.currentPiece = rotated

	// If the rotation causes a collision, revert back
	if g.checkCollision() {
		g.currentPiece = oldPiece
	}
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	rand.Seed(time.Now().UnixNano())
	game := NewGame()
	game.spawnPiece()

	// Game loop
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	gameOver := false
	for !gameOver {
		game.draw()

		select {
		case <-ticker.C:
			if !game.moveDown() {
				gameOver = true
			}
		case ev := <-eventQueue:
			if ev.Type == termbox.EventKey {
				switch ev.Key {
				case termbox.KeyArrowLeft:
					game.moveLeft()
				case termbox.KeyArrowRight:
					game.moveRight()
				case termbox.KeyArrowDown:
					game.moveDown()
				case termbox.KeyArrowUp:
					game.rotate()
				case termbox.KeyEsc:
					return
				}
			} else if ev.Type == termbox.EventError {
				panic(ev.Err)
			}
		}
	}

	// Display game over message
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	gameOverMsg := "Game Over! Final Score: " + fmt.Sprint(game.score)
	for i, char := range gameOverMsg {
		termbox.SetCell(i, boardHeight/2, char, termbox.ColorRed, termbox.ColorDefault)
	}
	termbox.Flush()
	time.Sleep(2 * time.Second) // Show game over message for 2 seconds
}
