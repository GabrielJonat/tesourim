package main

import (
	"fmt"
	"example/tesourim/utils"
	"image/color"
	"log"
	"math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"embed"
)

//go:embed assets/fonts
var fontFS embed.FS

const (
	gridWidth    = 600
	gridHeight   = 600
	playing = iota
	won
	lost
	memorizing = iota
)

// Constantes para o inimigo
const (
	bulletSpeed = 0.25
	enemyY     = -1.6
)

// Enemy representa o inimigo que se move e atira
type Enemy struct {
	x             float64
	errAcum       float64
	prevErr       float64
	bullets       []Bullet
	lastShotTimer int
	alive         bool
}

// Bullet representa um projétil
type Bullet struct {
	x, y    float64
	dx, dy  float64
	active  bool
	reflected bool
}

// NewEnemy cria um novo inimigo
func NewEnemy() *Enemy {
	return &Enemy{
		x:             float64(gridSize / 2),
		errAcum:       0,
		prevErr:       0,
		bullets:       make([]Bullet, 0),
		lastShotTimer: 0,
		alive:         true,
	}
}

// calculatePID calcula o movimento baseado no controle PID
func (e *Enemy) calculatePID(setpoint, currentPos float64) float64 {
	var kp, ki, kd float64

	// Ajusta as constantes do PID baseado na dificuldade
	switch dificulty {
	case 1: // Fácil - apenas proporcional
		kp, ki, kd = 0.3, 0, 0
	case 2: // Médio - proporcional + integral
		kp, ki, kd = 0.3, 0.1, 0
	case 3: // Difícil - PID completo
		kp, ki, kd = 0.3, 0.1, 0.05
	}

	err := setpoint - currentPos
	e.errAcum += err
	derivative := err - e.prevErr

	output := (kp * err) + (ki * e.errAcum) + (kd * derivative)
	e.prevErr = err

	// Limita a saída para movimento suave
	if output > 1 {
		output = 1
	} else if output < -1 {
		output = -1
	}

	return output
}

// Update atualiza a posição do inimigo e seus projéteis
func (e *Enemy) Update(playerX, playerY int) {
	if !e.alive {
		return
	}

	// Atualiza posição do inimigo usando PID
	targetX := float64(playerX)
	movement := e.calculatePID(targetX, e.x)
	e.x += movement

	// Mantém o inimigo dentro dos limites do grid
	if e.x < 0 {
		e.x = 0
	} else if e.x >= float64(gridSize) {
		e.x = float64(gridSize - 1)
	}

	// Atualiza timer de tiro
	if e.lastShotTimer > 0 {
		e.lastShotTimer--
	}

	// Tenta atirar se estiver próximo ao alinhamento com o jogador
	if e.lastShotTimer == 0 && math.Abs(float64(playerX)-e.x) < 0.5 {
		// Adiciona novo projétil
		if utils.RussianRoulette(dificulty) {
			e.bullets = append(e.bullets, Bullet{
				x:      e.x,
				y:      float64(enemyY),
				dx:     0,
				dy:     1,
				active: true,
				reflected: false,
			})
			e.lastShotTimer = 60 * 1.5
		}
		if !utils.RussianRoulette(dificulty) {
			e.bullets = append(e.bullets, Bullet{
				x:      e.x,
				y:      float64(enemyY),
				dx:     0,
				dy:     1,
				active: false,
				reflected: false,
			})
			e.lastShotTimer = 60 * 1.5
		}
		
	}

	// Atualiza projéteis
	for i := range e.bullets {
		if e.bullets[i].active {
			e.bullets[i].y += e.bullets[i].dy * bulletSpeed

			// Verifica colisão com o inimigo para projéteis refletidos
			if e.bullets[i].reflected {
				enemyGridX := int(math.Round(e.x))
				enemyGridY := int(math.Round(float64(enemyY)))
				bulletGridX := int(math.Round(e.bullets[i].x))
				bulletGridY := int(math.Round(e.bullets[i].y))
				
				if bulletGridX == enemyGridX && bulletGridY == enemyGridY && e.alive {
					e.alive = false
					e.bullets[i].active = false
				}
			}

			// Desativa projéteis fora do grid
			if e.bullets[i].y >= float64(gridSize) || e.bullets[i].y <= float64(enemyY-1) {
				e.bullets[i].active = false
			}
		}
	}

	// Remove projéteis inativos
	activeBullets := make([]Bullet, 0)
	for _, bullet := range e.bullets {
		if bullet.active {
			activeBullets = append(activeBullets, bullet)
		}
	}
	e.bullets = activeBullets
}

// Draw desenha o inimigo e seus projéteis
func (e *Enemy) Draw(screen *ebiten.Image, offsetX, offsetY int) {
	// Desenha o inimigo apenas se estiver vivo
	if e.alive {
		enemyScreenX := float64(offsetX) + (e.x * float64(nodeSize))
		enemyScreenY := float64(offsetY) + (float64(enemyY) * float64(nodeSize))
		ebitenutil.DrawRect(screen, enemyScreenX, enemyScreenY, float64(nodeSize), float64(nodeSize), color.RGBA{255, 0, 0, 255})
	}

	// Desenha os projéteis
	for _, bullet := range e.bullets {
		if bullet.active {
			bulletScreenX := float64(offsetX) + (bullet.x * float64(nodeSize)) + float64(nodeSize)/2
			bulletScreenY := float64(offsetY) + (bullet.y * float64(nodeSize)) + float64(nodeSize)/2
			// Projéteis refletidos são azuis
			bulletColor := color.RGBA{255, 255, 0, 255}
			if bullet.reflected {
				bulletColor = color.RGBA{0, 0, 255, 255}
			}
			ebitenutil.DrawCircle(screen, bulletScreenX, bulletScreenY, 12, bulletColor)
		}
	}
}

func setup(L int) (int, map[int]bool){
	graph := utils.GenerateGraph(L)
	target := utils.GenerateTreasure(L) 
	// Mark trap nodes
	traps := utils.GenerateTraps(L, target, dificulty)
	start := 0
	reachable := false
	for i := start; i < L ; i++ {
		if utils.CanReach(graph, traps, i, target) {
			reachable = true
			break
		}
	}
	// Check if the target is reachable
	if reachable {
		return target, traps
	} else {
		return setup(L)
	}
}

func init() {
	// Carregar a fonte normal
	fontData, err := fontFS.ReadFile("assets/fonts/Mplus1-SemiBold.ttf")
	if err != nil {
		log.Fatal(err)
	}

	tt, err := opentype.Parse(fontData)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    30,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Carregar a fonte bold (usando a mesma fonte para bold por enquanto)
	mplusBoldFont = mplusNormalFont
}

var (
	initialTarget, initialTraps = setup(gridSize)
	initialFallenTraps = make(map[int]bool)
	mplusNormalFont            font.Face
	mplusBoldFont             font.Face
	gridSize     = 6 // Number of rows and columns in the grid
	nodeSize     = gridWidth / gridSize
	dificulty    = 1
	memorizeTime = 30 * 60 // 10 seconds in frames (60 FPS)
	gameTime   = 40 * 60
	lives = 2
	restart = false
	enemy = NewEnemy()
	rocks = make([]Rock, 0) // Lista de pedras ativas
)

func levelUp() {
	dificulty++
	if dificulty == 4 {
		dificulty = 1
		gameTime += 60 * 5
		gridSize++
		if gridSize % 2 == 0 {
			lives++
		}
	}

	nodeSize = gridWidth / gridSize
}

type Game struct{
	playerX    int     // Grid position X
	playerY    int     // Grid position Y
	gameState  int     // Current game state
	message    string  // Game message (win/lose)
	timer      int     // Timer for memorization phase
	showTraps  bool    // Whether to show traps and treasure
	gameTimer  int     // Timer for gameplay phase
	lives      int     // Number of lives
	rocks      int     // Number of rocks available
	aimX       int     // Aiming position X
	aimY       int     // Aiming position Y
	aiming     bool    // Whether player is currently aiming
}

// Rock represents a thrown rock
type Rock struct {
	x, y       float64
	targetX    int
	targetY    int
	active     bool
	revealed   map[int]bool // Nodes revealed by this rock
}

func NewGame() *Game {
	return &Game{
		playerX:    0,  // Start outside the grid
		playerY:    -1,   // At the first row level
		gameState:  memorizing,
		message:    fmt.Sprintf("Memorize em %d segundos!", memorizeTime/60),
		timer:      memorizeTime,
		showTraps:  true,
		gameTimer:  gameTime,
		lives:      lives,
		rocks:      5, // Começa com 3 pedras
		aimX:       0,
		aimY:       0,
		aiming:     false,
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	target, traps := initialTarget, initialTraps
	fallenTraps := initialFallenTraps
	// Get screen dimensions to center the grid
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	offsetX := (sw - gridWidth) / 2
	offsetY := (sh - gridHeight) / 2
	
	// Draw title and instructions
	face := basicfont.Face7x13
	title := "Tesourim"
	instructions := "Pressione ESC para sair | Use WASD/Setas para mover | QEZC para diagonais" + " | level: " + fmt.Sprintf("%d", gridSize - 5)
	
	// Calculate text position for center alignment
	titleBounds := font.MeasureString(mplusNormalFont, title)
	titleX := float64(sw/2 - titleBounds.Round()/2)
	
	// Draw texts
	text.Draw(screen, title, mplusNormalFont, int(titleX), offsetY-20, color.White)
	text.Draw(screen, instructions, face, offsetX, offsetY-5, color.White)

	if g.gameState == playing {
		timeLeft := fmt.Sprintf("Tempo: %d", g.gameTimer/60)
		text.Draw(screen, timeLeft, mplusBoldFont, sw-180, 40, color.White)
		
		// Desenha as vidas restantes
		lives := fmt.Sprintf("Vidas: %d", g.lives)
		text.Draw(screen, lives, mplusBoldFont, 30, 40, color.White)

		// Opcional: Desenha corações para representar as vidas
		for i := 0; i < g.lives; i++ {
			heartX := float64(60 + (i * 20))
			ebitenutil.DrawCircle(screen, heartX, 55, 8, color.RGBA{255, 0, 0, 255})
		}
	}
	// Draw grid
	for row := 0; row < gridSize; row++ {
		for col := 0; col < gridSize; col++ {
			// Invert the row to start from bottom
			invertedRow := (gridSize - 1) - row
			node := invertedRow*gridSize + col

			// Determine the color for this cell
			var clr color.Color
			if g.showTraps {
				if traps[node] {
					clr = color.RGBA{255, 0, 0, 255} // Red for traps
				} else if node == target {
					clr = color.RGBA{0, 255, 0, 255} // Green for the treasure
				}else {
					clr = color.RGBA{200, 200, 200, 255} // Gray for normal nodes
				}
			} else {
					clr = color.RGBA{200, 200, 200, 255} // Gray for all nodes when hidden
			}
			if fallenTraps[node] {
				clr = color.RGBA{128, 128, 128, 255} // grey for fallen traps
			}


			// Draw the rectangle for the node with offset for centering
			x := float64(offsetX + (col * nodeSize))
			y := float64(offsetY + (row * nodeSize))
			ebitenutil.DrawRect(screen, x, y, float64(nodeSize), float64(nodeSize), clr)

			// Draw grid lines
			ebitenutil.DrawLine(screen, x, y, x+float64(nodeSize), y, color.Black)
			ebitenutil.DrawLine(screen, x, y, x, y+float64(nodeSize), color.Black)
		}
	}
	// Draw the final border lines
	lastX := float64(offsetX + gridWidth)
	lastY := float64(offsetY + gridHeight)
	ebitenutil.DrawLine(screen, lastX, float64(offsetY), lastX, lastY, color.Black)
	ebitenutil.DrawLine(screen, float64(offsetX), lastY, lastX, lastY, color.Black)
	// Draw the player
	if g.playerX >= 0 && g.playerY >= -1 {
		playerScreenX := float64(offsetX + (g.playerX * nodeSize))
		playerScreenY := float64(offsetY + ((gridSize-1-g.playerY) * nodeSize))
		ebitenutil.DrawRect(screen, playerScreenX, playerScreenY, float64(nodeSize), float64(nodeSize), color.RGBA{0, 0, 255, 255})
	}
	
	// Draw aiming crosshair when in aiming mode
	if g.aiming {
		aimScreenX := float64(offsetX) + (float64(g.aimX) * float64(nodeSize)) + float64(nodeSize)/2
		aimScreenY := float64(offsetY) + (float64(gridSize-1-g.aimY) * float64(nodeSize)) + float64(nodeSize)/2
		
		// Draw crosshair
		ebitenutil.DrawLine(screen, aimScreenX-10, aimScreenY, aimScreenX+10, aimScreenY, color.RGBA{255, 0, 0, 255})
		ebitenutil.DrawLine(screen, aimScreenX, aimScreenY-10, aimScreenX, aimScreenY+10, color.RGBA{255, 0, 0, 255})
	}

	// Draw rocks counter
	if g.gameState == playing {
		rocks := fmt.Sprintf("Pedras: %d", g.rocks)
		text.Draw(screen, rocks, mplusBoldFont, 30, 90, color.White)

		for i := 0; i < g.rocks; i++ {
			rockX := float64(60 + (i * 20))
			ebitenutil.DrawCircle(screen, rockX, 105, 8, color.RGBA{128, 128, 128, 255})
		}
	}

	// Draw active rocks
	for _, rock := range rocks {
		if rock.active {
			rockScreenX := float64(offsetX) + (rock.x * float64(nodeSize)) + float64(nodeSize)/2
			rockScreenY := float64(offsetY) + (float64(gridSize-1)-rock.y * float64(nodeSize)) + float64(nodeSize)/2
			ebitenutil.DrawCircle(screen, rockScreenX, rockScreenY, 5, color.RGBA{139, 69, 19, 255})
		}
	}

	// Draw revealed nodes
	for _, rock := range rocks {
		for node := range rock.revealed {
			// Calcula a linha e coluna corretamente
			col := node % gridSize
			row := node / gridSize
			
			// Calcula as coordenadas na tela
			x := float64(offsetX) + (float64(col) * float64(nodeSize))
			y := float64(offsetY) + (float64(gridSize-1-row) * float64(nodeSize))
			
			if node == initialTarget {
				ebitenutil.DrawRect(screen, x, y, float64(nodeSize), float64(nodeSize), color.RGBA{0, 255, 0, 128})
			} else if initialTraps[node] {
				ebitenutil.DrawRect(screen, x, y, float64(nodeSize), float64(nodeSize), color.RGBA{255, 0, 0, 128})
			} else {
				ebitenutil.DrawRect(screen, x, y, float64(nodeSize), float64(nodeSize), color.RGBA{200, 200, 200, 128})
			}
		}
	}
	
	// Draw game state message if exists
	if g.message != "" {
		msgBounds := font.MeasureString(mplusNormalFont, g.message)
		msgX := float64(sw/2 - msgBounds.Round()/2)
		text.Draw(screen, g.message, mplusNormalFont, int(msgX), sh/2, color.RGBA{255, 0, 255, 255})
	}
	if g.gameState == playing {
	enemy.Draw(screen, offsetX, offsetY)
	}
}

// Update handles the game state (not needed here).
func (g *Game) Update() error {
	// Check if ESC key is pressed to exit the game
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	// Handle memorization phase
	if g.gameState == memorizing {
		g.timer--
		if g.timer <= 0 {
			g.gameState = playing
			g.showTraps = false
			g.message = ""
		} else {
			g.message = fmt.Sprintf("Memorize em %d segundos!   Pressione SPACE para avançar", g.timer/60)
			if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				g.gameState = playing
				g.showTraps = false
				g.message = ""
			}
		}
		return nil
	}

	// Only allow movement if the game is still playing
	if g.gameState == playing {
		// Atualiza o timer do jogo
		g.gameTimer--
		if g.gameTimer <= 0 {
			g.gameState = lost
			restart = true
			resetFallenTraps()
			g.message = "Tempo esgotado! Pressione R para tentar novamente"
			return nil
		}

		if g.lives == 0 {
			g.gameState = lost
			restart = true
			resetFallenTraps()
			g.message = "Atingido! Pressione R para tentar novamente"
			return nil
		}

		// Update enemy
		enemy.Update(g.playerX, g.playerY)

		// Check bullet collisions and handle reflection
		for i := range enemy.bullets {
			if enemy.bullets[i].active {
				bulletGridX := int(math.Round(enemy.bullets[i].x))
				bulletGridY := gridSize - 1 - int(math.Round(enemy.bullets[i].y))
				
				// Verifica se o jogador está tentando refletir o projétil
				if inpututil.IsKeyJustPressed(ebiten.KeyV) {
					// Verifica se o projétil está próximo o suficiente para ser refletido
					if bulletGridX == g.playerX && math.Abs(float64(bulletGridY-g.playerY)) <= 1 {
						enemy.bullets[i].dy = -1 // Inverte a direção do projétil
						enemy.bullets[i].reflected = true
						return nil
					}
				}

				// Colisão normal se não foi refletido
				if bulletGridX == g.playerX && bulletGridY == g.playerY && !enemy.bullets[i].reflected {
					g.lives--
					enemy.bullets[i].active = false
					return nil
				}
			}
		}

		// Handle rock throwing mechanics
		if g.gameState == playing {
			// Enter/exit aiming mode with R key
			if inpututil.IsKeyJustPressed(ebiten.KeyControl) && g.rocks > 0{
				g.aiming = !g.aiming
				g.aimX = g.playerX
				g.aimY = g.playerY
			}

			// Handle aiming
			if g.aiming {
				// Calculate distance from player
				dx := g.aimX - g.playerX
				dy := g.aimY - g.playerY
				distance := math.Sqrt(float64(dx*dx + dy*dy))
				maxDistance := 5.0
				if distance > maxDistance {
					angle := math.Atan2(float64(dy), float64(dx))
					g.aimX = g.playerX + int(maxDistance * math.Cos(angle))
					g.aimY = g.playerY + int(maxDistance * math.Sin(angle))
				}
				if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
					if g.aimX > 0 {
						g.aimX--
					}
				}
				if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
					if g.aimX < gridSize-1 {
						g.aimX++
					}
				}
				if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
					if g.aimY < gridSize-1 {
						g.aimY++
					}
				}
				if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
					if g.aimY > 0 {
						g.aimY--
					}
				}

				// Throw rock with space
				if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
					g.rocks--
					// Calcula o nó corretamente
					node := (g.aimY) * gridSize + g.aimX
					
					// Create new rock
					newRock := Rock{
						x:       float64(g.playerX * nodeSize),
						y:       float64((gridSize - 1 - g.playerY) * nodeSize),
						targetX: g.aimX * nodeSize,
						targetY: (gridSize - 1 - g.aimY) * nodeSize,
						active:  true,
						revealed: make(map[int]bool),
					}
					
					// Reveal the target node
					newRock.revealed[node] = true
					
					// Check if hit treasure
					if node == initialTarget {
						g.gameState = won
						g.message = "Você achou o tesouro! Pressione ENTER para continuar"
					}

					rocks = append(rocks, newRock)
					g.aiming = false
				}
			}

			// Update rock positions
			for i := range rocks {
				if rocks[i].active {
					// Calculate direction to target
					dx := float64(rocks[i].targetX) - rocks[i].x
					dy := float64(rocks[i].targetY) - rocks[i].y
					length := math.Sqrt(dx*dx + dy*dy)
					// Calculate distance and clamp to max 5 nodes
					maxDistance := float64(5 * nodeSize) // 5 nós de distância
					if length > maxDistance {
						dx = dx * maxDistance / length
						dy = dy * maxDistance / length
						length = maxDistance
					}
					// Normalize and move
					speed := 0.2
					rocks[i].x += (dx / length) * speed
					rocks[i].y += (dy / length) * speed
					// Check if rock reached target
					if length < 0.1 {
						rocks[i].active = false
					}
				}
			}
		}

		if !g.aiming {
				// Handle movement
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
				g.tryMove(-1, 0)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
				g.tryMove(1, 0)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
				g.tryMove(0, 1)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
				g.tryMove(0, -1)
			}
			// Diagonal movements
			if inpututil.IsKeyJustPressed(ebiten.KeyQ) { // Up-left
				g.tryMove(-1, 1)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyE) { // Up-right
				g.tryMove(1, 1)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyZ) { // Down-left
				g.tryMove(-1, -1)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyC) { // Down-right
				g.tryMove(1, -1)
			}
		}
	}
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.playerX = 0
			g.playerY = -1
			g.gameState = playing
			g.showTraps = false
			g.aiming = false
			g.message = ""
			if restart {
				g.gameTimer = gameTime
				enemy = NewEnemy()
				g.rocks = 5  // Reseta o número de pedras
				rocks = make([]Rock, 0) // Limpa a lista de pedras e nós revelados
				restart = false
				g.gameState = memorizing
				g.lives = lives
				g.timer = memorizeTime
				g.showTraps = true
				g.message = fmt.Sprintf("Memorize em %d segundos!", g.timer/60)
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			if g.gameState == won {
				levelUp()
				resetFallenTraps()
				// Generate new game layout
				initialTarget, initialTraps = setup(gridSize)
				// Reset game state
				g.playerX = 0
				g.playerY = -1
				g.gameState = memorizing
				g.gameTimer = gameTime
				g.timer = memorizeTime
				enemy = NewEnemy()
				g.lives = lives
				g.aiming = false
				g.rocks = 5 // Reseta o número de pedras
				rocks = make([]Rock, 0) // Limpa a lista de pedras e nós revelados
				g.showTraps = true
				g.message = fmt.Sprintf("Memorize em %d segundos!", g.timer/60)
			}
		}
	return nil
}

func (g *Game) tryMove(dx, dy int) {
	newX := g.playerX + dx
	newY := g.playerY + dy
	node := (newY)*gridSize + newX
	
	// Check if the move is valid (within or just outside grid)
	if newX >= 0 && newX < gridSize && newY >= -1 && newY < gridSize && !initialFallenTraps[node]{
		g.playerX = newX
		g.playerY = newY
		
		// Only check for collisions if player is inside the grid
		if newY >= 0 {
			// Calculate node index
			
			// Check for trap collision
			if initialTraps[node] {
				g.gameState = lost
				g.message = "Você perdeu! Pressione R para reiniciar"
				updateFallenTraps(node)
			}
			
			// Check for treasure collision
			if node == initialTarget {
				g.gameState = won
				g.aiming = false
				g.message = "Você ganhou! Pressione ENTER para avançar"
			}
		}
	}
}

// Layout sets the screen dimensions.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func updateFallenTraps(node int) {
	initialFallenTraps[node] = true
}

func resetFallenTraps() {
	initialFallenTraps = make(map[int]bool)
}

func main() {
	ebiten.SetWindowSize(gridWidth, gridHeight)
	ebiten.SetWindowTitle("Tesourim")
	ebiten.SetFullscreen(true)
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}