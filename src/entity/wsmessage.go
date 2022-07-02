package entity

type WebSocketReq struct {
	Version string        `json:"version"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

type WebSocketRep struct {
	Result string `json:"result"`
	Id     int    `json:"id"`
}
