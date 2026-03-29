package main

type Player struct {
	Name    string `json:"name"`
	Kills   int    `json:"kills"`
	Deaths  int    `json:"deaths"`
	Matches int    `json:"matches"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type LeaderboardEntry struct {
	Name string  `json:"name"`
	KD   float64 `json:"kd"`
}

type UpdatePlayer struct {
	Kills   *int `json:"kills"`
	Deaths  *int `json:"deaths"`
	Matches *int `json:"matches"`
}

var players = []Player{
	{Name: "Alex", Kills: 10, Deaths: 5, Matches: 10},
	{Name: "Max", Kills: 5, Deaths: 7, Matches: 10},
	{Name: "John", Kills: 20, Deaths: 2, Matches: 10},
}
