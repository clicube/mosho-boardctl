package pkg

type Env struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Brightness  float64 `json:"brightness"`
}

type IrData struct {
	Pattern  string
	Interval int
}
