package web

type Config struct {
	IP   string `json:"IP"`
	PORT string `json:"PORT"`

	TokenUsdesk string `json:"token"`
	IdApp       string `json:"id-app"`
}

type Form struct {
	InitData      string `json:"initData"`
	UserName      string `json:"-"`
	Name          string `json:"name"`
	IsEmergency   bool   `json:"is-emergency"`
	Place         string `json:"place"`
	GroupNumber   string `json:"group-number"`
	Department    string `json:"departament"`
	ReadyProblem  string `json:"ready-problem"`
	CustomProblem string `json:"custom-problem"`
}
