package entity

type WebSocketReq struct {
	Version string        `json:"version"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

type WebSocketRep struct {
	Exception error  `json:"exception"`
	Message   string `json:"message"`
	Id        int    `json:"id"`
}
