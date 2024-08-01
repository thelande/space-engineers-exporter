package client

import "errors"

var ErrNoKeySpecified = errors.New("no secret key was specified")
var ErrNon2XXResponse = errors.New("received non 2XX status code")

type BaseResponse struct {
	Meta struct {
		ApiVersion string  `json:"apiVersion"`
		QueryTime  float64 `json:"queryTime"`
	} `json:"meta"`
}

type PingResponseData struct {
	Result string `json:"Result"`
}

type PingResponse struct {
	BaseResponse
	Data PingResponseData `json:"data"`
}

type ServerResponseData struct {
	IsReady           bool    `json:"IsReady"`
	PirateUsedPCU     uint    `json:"PirateUsedPCU"`
	Players           int     `json:"Players"`
	ServerId          uint64  `json:"ServerId"`
	ServerName        string  `json:"ServerName"`
	SimSpeed          float64 `json:"SimSpeed"`
	SimulationCpuLoad float64 `json:"SimulationCpuLoad"`
	TotalTime         uint    `json:"TotalTime"`
	UsedPCU           uint    `json:"UsedPCU"`
	Version           string  `json:"Version"`
	WorldName         string  `json:"WorldName"`
}

type ServerResponse struct {
	BaseResponse
	Data ServerResponseData `json:"data"`
}

type EntityPosition struct {
	X float64 `json:"X"`
	Y float64 `json:"Y"`
	Z float64 `json:"Z"`
}

type PlanetResponseData struct {
	DisplayName string         `json:"DisplayName"`
	EntityId    int64          `json:"EntityId"`
	Position    EntityPosition `json:"Position"`
}

type PlanetResponse struct {
	BaseResponse
	Data struct {
		Planets []PlanetResponseData `json:"Planets"`
	} `json:"data"`
}

type AsteroidResponseData struct {
	DisplayName string         `json:"DisplayName"`
	EntityId    int64          `json:"EntityId"`
	Position    EntityPosition `json:"Position"`
}

type AsteroidResponse struct {
	BaseResponse
	Data struct {
		Asteroids []AsteroidResponseData `json:"Asteroids"`
	} `json:"data"`
}

type GridResponseData struct {
	DisplayName      string         `json:"DisplayName"`
	EntityId         int64          `json:"EntityId"`
	GridSize         string         `json:"GridSize"`
	BlocksCount      uint           `json:"BlocksCount"`
	Mass             float64        `json:"Mass"`
	Position         EntityPosition `json:"Position"`
	LinearSpeed      float64        `json:"LinearSpeed"`
	DistanceToPlayer float64        `json:"DistanceToPlayer"`
	OwnerSteamId     uint64         `json:"OwnerSteamId"`
	OwnerDisplayName string         `json:"OwnerDisplayName"`
	IsPowered        bool           `json:"IsPowered"`
	PCU              uint           `json:"PCU"`
}

type GridResponse struct {
	BaseResponse
	Data struct {
		Grids []GridResponseData `json:"Grids"`
	} `json:"data"`
}

type PlayerResponseData struct {
	SteamID     uint64 `json:"SteamID"`
	DisplayName string `json:"DisplayName"`
}

type BannedPlayersResponse struct {
	BaseResponse
	Data struct {
		BannedPlayers []PlayerResponseData `json:"BannedPlayers"`
	} `json:"data"`
}

type KickedPlayersResponse struct {
	BaseResponse
	Data struct {
		KickedPlayers []PlayerResponseData `json:"KickedPlayers"`
	} `json:"data"`
}

type CheaterResponseData struct {
	Explanation    string `json:"Explanation"`
	Id             int    `json:"Id"`
	Name           string `json:"Name"`
	PlayerId       uint64 `json:"PlayerId"`
	ServerDateTime string `json:"ServerDateTime"`
}

type CheatersResponse struct {
	BaseResponse
	Data struct {
		Cheaters []CheaterResponseData `json:"Cheaters"`
	} `json:"data"`
}

// Union of all response types
type Response interface {
	PingResponse |
		ServerResponse |
		PlanetResponse |
		AsteroidResponse |
		GridResponse |
		BannedPlayersResponse |
		KickedPlayersResponse |
		CheatersResponse
}
