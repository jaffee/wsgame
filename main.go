package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

const (
	XPos = 64
)

type Command int

//go:generate stringer -type=Command

const (
	Nothing Command = iota
	Up
	Down
	Right
	Left
	Boost
	NewPlayer
)

type ComWrapper struct {
	Com Command
}

type GameStatus int

const (
	New GameStatus = iota
	Playing
)

type Scene struct {
	Player Player
	Scene  []SceneNode
}

type SceneNode struct {
	Texture string
	X       int
	Y       int
	Width   int
	Height  int
	ID      int
}

type Dude struct {
	X       float64
	Y       float64
	Texture string
	Vy      float64
	Width   int
	Height  int
	ID      int
}

func (d *Dude) String() string {
	return fmt.Sprintf("X:%v Y:%v Texture:%v Vy:%v Width:%v Height:%v", d.X, d.Y, d.Texture, d.Vy, d.Width, d.Height)

}

type PlayerID int

type GameCommand struct {
	Com      Command
	Conn     *websocket.Conn
	PlayerID PlayerID
}

type GameState struct {
	status  GameStatus
	players map[PlayerID]*Player
	dudes   map[PlayerID]*Dude
}

func NewGameState() *GameState {
	return &GameState{
		status:  New,
		players: make(map[PlayerID]*Player),
		dudes:   make(map[PlayerID]*Dude),
	}
}

func (gs *GameState) AddPlayer(conn *websocket.Conn, p PlayerID) {
	_, ok := gs.players[p]
	if ok {
		panic("already have this player, what are you doing")
	}
	gs.players[p] = &Player{
		Conn: conn,
		ID:   p,
	}
	gs.dudes[p] = &Dude{
		X:       XPos,
		Y:       120.0,
		Vy:      0.0,
		Texture: "images/red.png",
		Width:   16,
		Height:  16,
		ID:      int(p),
	}
	if len(gs.players) >= 2 {
		gs.dudes[p].Texture = "images/blue.png"
	}
}

type Player struct {
	Conn *websocket.Conn
	ID   PlayerID
}

func GameLoop(gc <-chan GameCommand) {
	state := NewGameState()
	tick := time.Tick(time.Millisecond * 16) // 60 FPS
	for range tick {
		coms := readCommands(gc)
		if len(coms) > 0 {
			log.Println("State before: ", state)
		}
		updateGameState(coms, state)
		if len(coms) > 0 {
			log.Println("State after: ", state)
		}
		scenes := createScenes(state)
		sendScenes(scenes)
	}
}

// readCommands drains the gc channel and returns all commands as soon as the channel is empty
func readCommands(gc <-chan GameCommand) []GameCommand {
	coms := make([]GameCommand, 0)
	for {
		select {
		case com := <-gc:
			coms = append(coms, com)
		default:
			return coms
		}
	}
}

func updateGameState(coms []GameCommand, state *GameState) {
	for _, com := range coms {
		updateState(com, state)
	}
}

func updateState(com GameCommand, state *GameState) {
	if com.Com == NewPlayer {
		state.AddPlayer(com.Conn, com.PlayerID)
		return
	}
	d, ok := state.dudes[com.PlayerID]
	if !ok {
		panic("Shouldn't be possible not to have a dude here")
	}
	log.Println("before: ", d)
	switch com.Com {
	case Up:
		log.Println("UP!!")
		d.Y -= 10
	case Down:
		d.Y += 10
	case Right:
		log.Printf("right")
		d.X += 10
	case Left:
		log.Printf("left")
		d.X -= 10
	}
	log.Println("afterd: ", d)
	log.Println("after: ", state.dudes[com.PlayerID])
}

func createScenes(state *GameState) []Scene {
	ret := make([]Scene, len(state.players))
	playNum := 0
	for pid, _ := range state.dudes {
		player := state.players[pid]
		ret[playNum] = sceneForPlayer(state, player)
		playNum += 1
	}
	return ret
}

func sceneForPlayer(state *GameState, p *Player) Scene {
	playerDude := state.dudes[p.ID]
	scene := make([]SceneNode, 0)
	for _, dude := range state.dudes {
		sn := SceneNode{
			X:       int(dude.X - playerDude.X + XPos),
			Y:       int(dude.Y),
			Texture: dude.Texture,
			Width:   dude.Width,
			Height:  dude.Height,
			ID:      dude.ID,
		}
		scene = append(scene, sn)
	}
	return Scene{
		Scene:  scene,
		Player: *p,
	}
}

func sendScenes(scenes []Scene) {
	for _, s := range scenes {
		enc := json.NewEncoder(s.Player.Conn)
		err := enc.Encode(s.Scene)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	gameChan := make(chan GameCommand, 100)
	go GameLoop(gameChan)
	http.Handle("/game", GameHandler(gameChan))
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func GameHandler(gameChan chan<- GameCommand) websocket.Handler {
	return func(ws *websocket.Conn) {
		plid := PlayerID(rand.Intn(10))
		gameChan <- GameCommand{
			Com:      NewPlayer,
			Conn:     ws,
			PlayerID: plid,
		}
		for {
			dec := json.NewDecoder(ws)
			var c ComWrapper
			err := dec.Decode(&c)
			if err != nil {
				time.Sleep(time.Millisecond * 4)
				continue
			}
			if c.Com != Nothing {
				log.Printf("Received %v for plid: %v", c.Com.String(), plid)
				gameChan <- GameCommand{
					Com:      c.Com,
					Conn:     ws,
					PlayerID: plid,
				}
			}
		}
	}
}
