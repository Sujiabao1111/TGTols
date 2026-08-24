package dtos

type Todo struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type RepJsonEcho struct {
	Me string `json:"me"`
}

type RespJsonEchoBack struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Ip      string `json:"Ip"`
}
