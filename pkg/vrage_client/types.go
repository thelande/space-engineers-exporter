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
	Players           uint    `json:"Players"`
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
